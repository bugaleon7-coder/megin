package router

import (
	"errors"
	"fmt"
	"io"
	"megin/pkg/context/api"
	"megin/pkg/errs"
	"net/http"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RouteInfo 路由元数据
type RouteInfo struct {
	Method       string
	Path         string      // 完整路径（含分组前缀）
	ReqType      interface{} // 请求结构体零值
	RespType     interface{} // 响应结构体零值
	HandlerName  string      // 处理函数短名（不含包路径），用于提取注释
	ReceiverType string      // 方法接收者类型名，用于提取 @Tag
}

// RouteCollector 路由收集器接口
type RouteCollector interface {
	AddRoute(method, path string, handler gin.HandlerFunc, reqType, respType interface{})
}

// handlerNameSetter 内部接口，供泛型注册函数通过反射写入 HandlerName，不对外暴露
type handlerNameSetter interface {
	setLastHandlerName(name string)
	setLastReceiverType(name string)
}

// RouteRegistry 根路由注册器（同时也是根 Group）
type RouteRegistry struct {
	Engine    *gin.Engine
	routes    []RouteInfo
	rootGroup *gin.RouterGroup
}

// NewRouteRegistry 创建注册器
func NewRouteRegistry(engine *gin.Engine) *RouteRegistry {
	rootGroup := engine.Group("/")
	return &RouteRegistry{
		Engine:    engine,
		routes:    make([]RouteInfo, 0),
		rootGroup: rootGroup,
	}
}

// AddRoute 实现 RouteCollector 接口（根路由注册）
func (r *RouteRegistry) AddRoute(method, path string, handler gin.HandlerFunc, reqType, respType interface{}) {
	r.Engine.Handle(method, path, handler)
	r.routes = append(r.routes, RouteInfo{
		Method: method, Path: path, ReqType: reqType, RespType: respType,
	})
}

func (r *RouteRegistry) setLastHandlerName(name string) {
	if len(r.routes) > 0 {
		r.routes[len(r.routes)-1].HandlerName = name
	}
}

func (r *RouteRegistry) setLastReceiverType(name string) {
	if len(r.routes) > 0 {
		r.routes[len(r.routes)-1].ReceiverType = name
	}
}

// Use 全局中间件（作用于所有路由）
func (r *RouteRegistry) Use(handlers ...gin.HandlerFunc) {
	r.Engine.Use(handlers...)
}

// Group 创建子分组
func (r *RouteRegistry) Group(relativePath string, handlers ...gin.HandlerFunc) *RouteGroup {
	rg := r.Engine.Group(relativePath, handlers...)
	return &RouteGroup{rg: rg, registry: r}
}

// Routes 返回所有路由元数据（供文档生成）
func (r *RouteRegistry) Routes() []RouteInfo {
	return r.routes
}

// RouteGroup 路由分组
type RouteGroup struct {
	rg       *gin.RouterGroup
	registry *RouteRegistry
}

// Use 分组中间件
func (g *RouteGroup) Use(handlers ...gin.HandlerFunc) {
	g.rg.Use(handlers...)
}

// Group 基于当前分组创建子分组
func (g *RouteGroup) Group(relativePath string, handlers ...gin.HandlerFunc) *RouteGroup {
	rg := g.rg.Group(relativePath, handlers...)
	return &RouteGroup{rg: rg, registry: g.registry}
}

// AddRoute 实现 RouteCollector 接口（分组路由注册）
func (g *RouteGroup) AddRoute(method, path string, handler gin.HandlerFunc, reqType, respType interface{}) {
	g.rg.Handle(method, path, handler)
	fullPath := g.rg.BasePath() + path
	g.registry.routes = append(g.registry.routes, RouteInfo{
		Method: method, Path: fullPath, ReqType: reqType, RespType: respType,
	})
}

func (g *RouteGroup) setLastHandlerName(name string) {
	g.registry.setLastHandlerName(name)
}

func (g *RouteGroup) setLastReceiverType(name string) {
	g.registry.setLastReceiverType(name)
}

// ---------- 泛型路由注册函数 ----------

// GET 注册 GET 接口，并按需应用请求策略。
func GET[Req any, Resp any](c RouteCollector, relativePath string, handler func(*api.Context, Req) (Resp, error), options ...api.RequestOption) {
	fn := WrapHandlerForParams(handler, options...)
	c.AddRoute("GET", relativePath, fn, zero[Req](), zero[Resp]())
	setHandlerName(c, handler)
}

// POST 注册 POST 接口，并按需应用请求策略。
func POST[Req any, Resp any](c RouteCollector, relativePath string, handler func(*api.Context, Req) (Resp, error), options ...api.RequestOption) {
	fn := WrapHandlerWithBody(handler, options...)
	c.AddRoute("POST", relativePath, fn, zero[Req](), zero[Resp]())
	setHandlerName(c, handler)
}

func PUT[Req any, Resp any](c RouteCollector, relativePath string,
	handler func(*api.Context, Req) (Resp, error)) {
	fn := WrapHandlerWithBody(handler)
	c.AddRoute("PUT", relativePath, fn, zero[Req](), zero[Resp]())
	setHandlerName(c, handler)
}

func DELETE[Req any, Resp any](c RouteCollector, relativePath string,
	handler func(*api.Context, Req) (Resp, error)) {
	fn := WrapHandlerForParams(handler)
	c.AddRoute("DELETE", relativePath, fn, zero[Req](), zero[Resp]())
	setHandlerName(c, handler)
}

// DELETEWithBody 用于兼容以 JSON body 传参的 DELETE 接口。
func DELETEWithBody[Req any, Resp any](c RouteCollector, relativePath string,
	handler func(*api.Context, Req) (Resp, error)) {
	fn := WrapHandlerWithBody(handler)
	c.AddRoute("DELETE", relativePath, fn, zero[Req](), zero[Resp]())
	setHandlerName(c, handler)
}

// setHandlerName 通过反射获取原始 handler 的函数短名，写入最后一条路由
var receiverTypeRe = regexp.MustCompile(`\(\*?(\w+)\)`)

func setHandlerName[Req any, Resp any](c RouteCollector, handler func(*api.Context, Req) (Resp, error)) {
	ns, ok := c.(handlerNameSetter)
	if !ok {
		return
	}
	pc := reflect.ValueOf(handler).Pointer()
	fullName := runtime.FuncForPC(pc).Name()
	// 提取接收者类型名："example.com/m/handler.(*User).GetUser-fm" → "User"
	if m := receiverTypeRe.FindStringSubmatch(fullName); len(m) > 1 {
		ns.setLastReceiverType(m[1])
	}
	// fullName 形如 "example.com/m/handler.(*User).GetUser-fm"，取最后一段并去掉 -fm 后缀
	if idx := strings.LastIndex(fullName, "."); idx >= 0 {
		fullName = fullName[idx+1:]
	}
	fullName = strings.TrimSuffix(fullName, "-fm")
	ns.setLastHandlerName(fullName)
}

// ---------- 辅助函数 ----------

func zero[T any]() T { var v T; return v }

// WrapHandlerWithBody 绑定 JSON Body，并按需执行请求策略。
func WrapHandlerWithBody[Req any, Resp any](fn func(*api.Context, Req) (Resp, error), definitions ...api.RequestOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		target := requestBindingTarget(&req)
		if err := c.ShouldBindJSON(target); err != nil {
			if !errors.Is(err, io.EOF) {
				zhError := parseBindingError(err, req)

				respError := &errs.NormalError{
					Code:    4000,
					Message: err.Error(),
					Err:     err,
				}
				errmsg := api.Failed[error](respError, "绑定参数异常:"+zhError)
				c.JSON(http.StatusOK, errmsg)
				return
			}
			if err := binding.Validator.ValidateStruct(target); err != nil {
				errmsg := api.Failed[error](err, "校验参数异常")
				c.JSON(http.StatusOK, errmsg)
				return
			}
		}

		ctx, err := api.NewContext(c)
		if err != nil {
			c.JSON(http.StatusOK, api.Failed[error](err))
			return
		}
		opts := api.NewOptions(ctx, req, definitions...)
		resp, err := api.Execute(opts, func() (Resp, error) { return fn(ctx, req) })

		if err != nil {
			errmsg := api.Failed[error](err)
			c.JSON(http.StatusOK, errmsg)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

// 获取字段的 tag，优先级：json > form > 字段名
// 获取字段的 tag，优先级：json > form > 字段名
func getFieldTag(structType reflect.Type, fieldName string, tags ...string) string {
	// 通过字段名获取 StructField
	field, ok := structType.FieldByName(fieldName)
	if !ok {
		return fieldName
	}

	// 按优先级获取 tag
	for _, tag := range tags {
		name := strings.SplitN(field.Tag.Get(tag), ",", 2)[0]
		if name != "" && name != "-" {
			return name
		}
	}
	return fieldName
}

func parseBindingError(err error, obj interface{}) string {

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		// 获取结构体类型
		t := reflect.TypeOf(obj)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		for _, e := range validationErrors {
			// 通过字段名获取结构体字段
			fieldName := getFieldTag(t, e.Field(), "json", "form")

			switch e.Tag() {
			case "required":
				return fieldName + " 不能为空"
			case "min":
				return fieldName + " 长度不能小于 " + e.Param() + " 个字符"
			case "max":
				return fieldName + " 长度不能大于 " + e.Param() + " 个字符"
			default:
				return e.Tag() + " 验证失败"
			}
		}
	}
	return err.Error()
}

// WrapHandlerForParams 绑定 path/form 参数，并按需执行请求策略。
func WrapHandlerForParams[Req any, Resp any](fn func(*api.Context, Req) (Resp, error), definitions ...api.RequestOption) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req Req
		if err := BindParams(c, requestBindingTarget(&req)); err != nil {
			errmsg := api.Failed[error](err, "参数异常")
			c.JSON(http.StatusOK, errmsg)
			return
		}

		ctx, err := api.NewContext(c)
		if err != nil {
			c.JSON(http.StatusOK, api.Failed[error](err))
			return
		}

		opts := api.NewOptions(ctx, req, definitions...)
		resp, err := api.Execute(opts, func() (Resp, error) { return fn(ctx, req) })
		if err != nil {
			errmsg := api.Failed[error](err)
			c.JSON(http.StatusOK, errmsg)
			return
		}
		c.JSON(http.StatusOK, resp)
	}
}

func requestBindingTarget[Req any](req *Req) interface{} {
	v := reflect.ValueOf(req).Elem()
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return v.Interface()
	}
	return req
}

// BindParams binds query and path parameters, including embedded structs,
// pointers, slices and the standard types supported by Gin's form mapper.
func BindParams(c *gin.Context, ptr interface{}) error {
	if err := binding.MapFormWithTag(ptr, c.Request.URL.Query(), "form"); err != nil {
		return err
	}
	if err := bindPathFields(reflect.ValueOf(ptr), c); err != nil {
		return err
	}
	return binding.Validator.ValidateStruct(ptr)
}

func bindPathFields(v reflect.Value, c *gin.Context) error {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return nil
	}

	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fv := v.Field(i)
		if field.Anonymous {
			if err := bindPathFields(fv, c); err != nil {
				return err
			}
		}
		tag := field.Tag.Get("path")
		if tag == "" {
			continue
		}
		val := c.Param(tag)
		if val == "" && strings.Contains(field.Tag.Get("binding"), "required") {
			return fmt.Errorf("path parameter %s is required", tag)
		}
		if val != "" {
			if err := setField(fv, val); err != nil {
				return fmt.Errorf("field %s: %w", field.Name, err)
			}
		}
	}
	return nil
}

func setField(fv reflect.Value, val string) error {
	switch fv.Kind() {
	case reflect.String:
		fv.SetString(val)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
		fv.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return err
		}
		fv.SetUint(u)
	case reflect.Bool:
		b, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		fv.SetBool(b)
	default:
		return fmt.Errorf("unsupported type %s", fv.Type())
	}
	return nil
}
