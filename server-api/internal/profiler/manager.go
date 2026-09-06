package profiler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	runtimepprof "runtime/pprof"
	runtimetrace "runtime/trace"
	"sync"
	"time"

	googlepprof "github.com/google/pprof/profile"
)

const (
	defaultAddr       = "127.0.0.1:6060"
	shutdownTimeout   = 3 * time.Second
	maxProfileSeconds = 600
)

// Status 描述当前进程的 pprof 运行状态。
type Status struct {
	Enabled       bool   `json:"enabled"`
	Address       string `json:"address"`
	LastError     string `json:"last_error"`
	Goroutines    int    `json:"goroutines"`
	HeapAlloc     uint64 `json:"heap_alloc"`
	HeapInUse     uint64 `json:"heap_in_use"`
	StackInUse    uint64 `json:"stack_in_use"`
	GCCount       uint32 `json:"gc_count"`
	LastGCPauseNS uint64 `json:"last_gc_pause_ns"`
}

// FlameNode 是调用栈图节点，Value 的单位由所属 Profile 的 SampleUnit 决定。
type FlameNode struct {
	Name     string       `json:"name"`
	Value    int64        `json:"value"`
	Children []*FlameNode `json:"children,omitempty"`
}

// CPUProfile 是一次运行时 Profile 转换后的调用栈图数据。
type CPUProfile struct {
	ProfileType     string     `json:"profile_type"`
	DurationSeconds int        `json:"duration_seconds"`
	SampleUnit      string     `json:"sample_unit"`
	Total           int64      `json:"total"`
	Root            *FlameNode `json:"root"`
}

type CaptureResult struct {
	ProfileType string
	Profile     *CPUProfile
	Raw         []byte
	RawExt      string
	Error       string
}

type Manager struct {
	mu        sync.RWMutex
	profileMu sync.Mutex
	addr      string
	enabled   bool
	server    *http.Server
	listener  net.Listener
	lastError string
}

var defaultManager = &Manager{addr: defaultAddr}

func Default() *Manager { return defaultManager }

// Configure 设置监听地址，并按配置决定是否在启动时开启。
func Configure(addr string, enabled bool) error {
	if addr == "" {
		addr = defaultAddr
	}
	defaultManager.mu.Lock()
	defaultManager.addr = addr
	defaultManager.mu.Unlock()
	return defaultManager.SetEnabled(enabled)
}

// SetEnabled 动态启停只监听本机的 pprof HTTP 服务。
func (m *Manager) SetEnabled(enabled bool) error {
	if enabled {
		return m.start()
	}
	return m.stop()
}

func (m *Manager) start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.enabled {
		return nil
	}
	listener, err := net.Listen("tcp", m.addr)
	if err != nil {
		m.lastError = err.Error()
		return fmt.Errorf("启动 pprof 失败: %w", err)
	}
	server := &http.Server{Handler: http.DefaultServeMux, ReadHeaderTimeout: 5 * time.Second}
	m.listener = listener
	m.server = server
	m.enabled = true
	m.lastError = ""
	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			m.mu.Lock()
			m.enabled = false
			m.lastError = serveErr.Error()
			m.mu.Unlock()
		}
	}()
	return nil
}

func (m *Manager) stop() error {
	m.mu.Lock()
	if !m.enabled {
		m.lastError = ""
		m.mu.Unlock()
		return nil
	}
	server := m.server
	m.enabled = false
	m.server = nil
	m.listener = nil
	m.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		m.mu.Lock()
		m.lastError = err.Error()
		m.mu.Unlock()
		return fmt.Errorf("关闭 pprof 失败: %w", err)
	}
	m.mu.Lock()
	m.lastError = ""
	m.mu.Unlock()
	return nil
}

func (m *Manager) Status() Status {
	m.mu.RLock()
	status := Status{Enabled: m.enabled, Address: m.addr, LastError: m.lastError}
	m.mu.RUnlock()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	status.Goroutines = runtime.NumGoroutine()
	status.HeapAlloc = mem.HeapAlloc
	status.HeapInUse = mem.HeapInuse
	status.StackInUse = mem.StackInuse
	status.GCCount = mem.NumGC
	status.LastGCPauseNS = mem.PauseNs[(mem.NumGC+255)%256]
	return status
}

// Capture 采集指定类型的运行时 profile；Trace 只返回原始时间线文件。
func (m *Manager) Capture(ctx context.Context, profileType string, seconds int) (CaptureResult, error) {
	m.mu.RLock()
	enabled := m.enabled
	m.mu.RUnlock()
	if !enabled {
		return CaptureResult{}, errors.New("请先开启 pprof")
	}
	if seconds < 0 || seconds > maxProfileSeconds {
		return CaptureResult{}, fmt.Errorf("采样时长必须在 0-%d 秒之间", maxProfileSeconds)
	}
	if !m.profileMu.TryLock() {
		return CaptureResult{}, errors.New("已有采样任务正在执行")
	}
	defer m.profileMu.Unlock()

	if profileType == "trace" {
		raw, err := captureTrace(ctx, seconds)
		if err != nil {
			return CaptureResult{}, err
		}
		return CaptureResult{ProfileType: profileType, Raw: raw, RawExt: ".trace"}, nil
	}

	var raw []byte
	var err error
	switch profileType {
	case "cpu":
		raw, err = captureCPU(ctx, seconds)
	case "mutex":
		previous := runtime.SetMutexProfileFraction(5)
		defer runtime.SetMutexProfileFraction(previous)
		err = waitForCapture(ctx, seconds)
		if err == nil {
			raw, err = captureLookup("mutex")
		}
	case "block":
		runtime.SetBlockProfileRate(1)
		defer runtime.SetBlockProfileRate(0)
		err = waitForCapture(ctx, seconds)
		if err == nil {
			raw, err = captureLookup("block")
		}
	case "heap", "allocs", "goroutine", "threadcreate":
		if profileType == "heap" {
			runtime.GC()
		}
		raw, err = captureLookup(profileType)
	default:
		return CaptureResult{}, fmt.Errorf("不支持的 profile 类型: %s", profileType)
	}
	if err != nil {
		return CaptureResult{}, err
	}
	parsed, err := googlepprof.ParseData(raw)
	if err != nil {
		return CaptureResult{}, fmt.Errorf("解析 %s profile 失败: %w", profileType, err)
	}
	flame, err := buildProfile(parsed, seconds, profileType)
	if err != nil {
		return CaptureResult{}, err
	}
	return CaptureResult{ProfileType: profileType, Profile: &flame, Raw: raw, RawExt: ".pb.gz"}, nil
}

// CaptureAll 在同一个采样窗口内收集所有运行时诊断数据。
// CPU、Trace、Mutex 与 Block 覆盖整个时间窗口，其余类型在窗口结束时抓取快照。
func (m *Manager) CaptureAll(ctx context.Context, seconds int) ([]CaptureResult, error) {
	m.mu.RLock()
	enabled := m.enabled
	m.mu.RUnlock()
	if !enabled {
		return nil, errors.New("请先开启 pprof")
	}
	if seconds < 1 || seconds > maxProfileSeconds {
		return nil, fmt.Errorf("采样时长必须在 1-%d 秒之间", maxProfileSeconds)
	}
	if !m.profileMu.TryLock() {
		return nil, errors.New("已有采样任务正在执行")
	}
	defer m.profileMu.Unlock()

	previousMutexRate := runtime.SetMutexProfileFraction(5)
	defer runtime.SetMutexProfileFraction(previousMutexRate)
	runtime.SetBlockProfileRate(1)
	defer runtime.SetBlockProfileRate(0)

	var traceBuffer bytes.Buffer
	if err := runtimetrace.Start(&traceBuffer); err != nil {
		return nil, fmt.Errorf("启动 Trace 采样失败: %w", err)
	}
	traceStarted := true
	defer func() {
		if traceStarted {
			runtimetrace.Stop()
		}
	}()

	var cpuBuffer bytes.Buffer
	if err := runtimepprof.StartCPUProfile(&cpuBuffer); err != nil {
		return nil, fmt.Errorf("启动 CPU 采样失败: %w", err)
	}
	cpuStarted := true
	defer func() {
		if cpuStarted {
			runtimepprof.StopCPUProfile()
		}
	}()

	waitErr := waitForCapture(ctx, seconds)
	runtimepprof.StopCPUProfile()
	cpuStarted = false
	runtimetrace.Stop()
	traceStarted = false
	if waitErr != nil {
		return nil, waitErr
	}

	results := []CaptureResult{
		buildCaptureResult("cpu", cpuBuffer.Bytes(), ".pb.gz", seconds),
		{ProfileType: "trace", Raw: traceBuffer.Bytes(), RawExt: ".trace"},
	}
	runtime.GC()
	for _, profileType := range []string{"heap", "allocs", "goroutine", "mutex", "block", "threadcreate"} {
		raw, err := captureLookup(profileType)
		if err != nil {
			results = append(results, CaptureResult{ProfileType: profileType, Error: err.Error()})
			continue
		}
		results = append(results, buildCaptureResult(profileType, raw, ".pb.gz", seconds))
	}
	return results, nil
}

func buildCaptureResult(profileType string, raw []byte, rawExt string, seconds int) CaptureResult {
	result := CaptureResult{ProfileType: profileType, Raw: raw, RawExt: rawExt}
	parsed, err := googlepprof.ParseData(raw)
	if err != nil {
		result.Error = fmt.Sprintf("解析 %s profile 失败: %v", profileType, err)
		return result
	}
	profile, err := buildProfile(parsed, seconds, profileType)
	if err != nil {
		result.Error = err.Error()
		return result
	}
	result.Profile = &profile
	return result
}

func captureCPU(ctx context.Context, seconds int) ([]byte, error) {
	if seconds < 1 {
		return nil, errors.New("CPU 采样时长不能小于 1 秒")
	}
	var buffer bytes.Buffer
	if err := runtimepprof.StartCPUProfile(&buffer); err != nil {
		return nil, fmt.Errorf("启动 CPU 采样失败: %w", err)
	}
	err := waitForCapture(ctx, seconds)
	runtimepprof.StopCPUProfile()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func captureTrace(ctx context.Context, seconds int) ([]byte, error) {
	if seconds < 1 {
		return nil, errors.New("Trace 采样时长不能小于 1 秒")
	}
	var buffer bytes.Buffer
	if err := runtimetrace.Start(&buffer); err != nil {
		return nil, fmt.Errorf("启动 Trace 采样失败: %w", err)
	}
	err := waitForCapture(ctx, seconds)
	runtimetrace.Stop()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func captureLookup(name string) ([]byte, error) {
	profile := runtimepprof.Lookup(name)
	if profile == nil {
		return nil, fmt.Errorf("运行时不支持 %s profile", name)
	}
	var buffer bytes.Buffer
	if err := profile.WriteTo(&buffer, 0); err != nil {
		return nil, fmt.Errorf("采集 %s profile 失败: %w", name, err)
	}
	return buffer.Bytes(), nil
}

func waitForCapture(ctx context.Context, seconds int) error {
	timer := time.NewTimer(time.Duration(seconds) * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

type flameBuilder struct {
	node     *FlameNode
	children map[string]*flameBuilder
}

func buildCPUProfile(profile *googlepprof.Profile, seconds int) (CPUProfile, error) {
	return buildProfile(profile, seconds, "cpu")
}

func buildProfile(profile *googlepprof.Profile, seconds int, profileType string) (CPUProfile, error) {
	if len(profile.SampleType) == 0 {
		return CPUProfile{}, fmt.Errorf("%s profile 没有采样数据", profileType)
	}
	valueIndex := len(profile.SampleType) - 1
	wantedSample := map[string]string{
		"cpu": "cpu", "heap": "inuse_space", "allocs": "alloc_space",
		"goroutine": "goroutine", "mutex": "delay", "block": "delay", "threadcreate": "threadcreate",
	}[profileType]
	for index, sampleType := range profile.SampleType {
		if sampleType.Type == wantedSample {
			valueIndex = index
			break
		}
	}
	root := &flameBuilder{node: &FlameNode{Name: "全部"}, children: map[string]*flameBuilder{}}
	for _, sample := range profile.Sample {
		if valueIndex >= len(sample.Value) || sample.Value[valueIndex] <= 0 {
			continue
		}
		value := sample.Value[valueIndex]
		root.node.Value += value
		current := root
		for index := len(sample.Location) - 1; index >= 0; index-- {
			name := locationName(sample.Location[index])
			child := current.children[name]
			if child == nil {
				child = &flameBuilder{node: &FlameNode{Name: name}, children: map[string]*flameBuilder{}}
				current.children[name] = child
				current.node.Children = append(current.node.Children, child.node)
			}
			child.node.Value += value
			current = child
		}
	}
	if root.node.Value == 0 {
		return CPUProfile{}, fmt.Errorf("%s profile 没有可展示的采样数据", profileType)
	}
	return CPUProfile{
		ProfileType:     profileType,
		DurationSeconds: seconds,
		SampleUnit:      profile.SampleType[valueIndex].Unit,
		Total:           root.node.Value,
		Root:            root.node,
	}, nil
}

func locationName(location *googlepprof.Location) string {
	for _, line := range location.Line {
		if line.Function != nil && line.Function.Name != "" {
			return line.Function.Name
		}
	}
	return fmt.Sprintf("0x%x", location.Address)
}
