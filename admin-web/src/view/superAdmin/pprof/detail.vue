<template>
  <div class="pprof-detail-page">
    <div class="gva-card detail-header">
      <div class="title-row">
        <div>
          <div class="page-title">Pprof 综合分析 · 记录 #{{ record.id || route.params.id }}</div>
          <div class="page-tip">本次记录包含同一采样窗口内的 CPU、内存、协程、锁、阻塞、线程创建与 Trace 数据。</div>
        </div>
        <el-button @click="router.back()">返回记录列表</el-button>
      </div>
      <el-descriptions v-if="record.id" :column="4" border class="record-summary">
        <el-descriptions-item label="状态">
          <el-tag :type="record.status === 'completed' ? 'success' : record.status === 'failed' ? 'danger' : 'warning'" effect="plain">
            {{ statusLabels[record.status] || record.status }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="采样时长">{{ formatSeconds(record.duration_seconds) }}</el-descriptions-item>
        <el-descriptions-item label="样本总量">{{ formatSample(record.sample_total, record.sample_unit) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatTime(record.finished_at) }}</el-descriptions-item>
      </el-descriptions>
    </div>

    <div class="gva-card chart-card" v-loading="loading">
      <el-tabs v-model="selectedProfile" class="profile-tabs" @tab-change="renderSelectedProfile">
        <el-tab-pane v-for="item in profileTypeOptions" :key="item.value" :name="item.value" :label="item.label" />
      </el-tabs>
      <el-tabs v-model="viewTab" type="card" class="view-tabs" @tab-change="renderSelectedProfile">
        <el-tab-pane label="火焰图" name="flame">
          <el-alert v-if="errorMessage" :title="errorMessage" :type="selectedProfile === 'trace' && record.status === 'completed' ? 'info' : 'warning'" :closable="false" show-icon />
          <template v-else>
            <div class="section-heading flame-heading">
              <span>调用栈火焰图</span>
              <small>从上往下是调用方向；宽度表示累计样本占比，使用底部滑块缩放，鼠标滚轮只滚动页面</small>
            </div>
            <div ref="chartElement" class="flame-chart" />
          </template>
        </el-tab-pane>
        <el-tab-pane label="性能解读" name="interpretation">
          <div class="interpretation-box">
            <div class="interpretation-title">{{ currentInterpretation.title }}</div>
            <div class="interpretation-text">{{ currentInterpretation.description }}</div>
            <div class="interpretation-tip">{{ currentInterpretation.tip }}</div>
          </div>
          <el-alert v-if="errorMessage" :title="errorMessage" :type="selectedProfile === 'trace' && record.status === 'completed' ? 'info' : 'warning'" :closable="false" show-icon />
          <div v-else class="top-section">
            <div class="section-heading">
              <span>性能影响最大的 Top 10 操作</span>
              <small>按自身开销排序，累计开销包含其下游调用</small>
            </div>
            <el-table :data="topOperations" border stripe size="small" class="top-table">
              <el-table-column type="index" label="#" width="52" />
              <el-table-column label="函数 / 操作" min-width="420">
                <template #default="{ row }">
                  <div class="operation-name">{{ row.name }}</div>
                  <div class="operation-path" :title="row.path">{{ row.path }}</div>
                </template>
              </el-table-column>
              <el-table-column label="自身开销" width="150">
                <template #default="{ row }">{{ formatSample(row.selfValue, currentProfile.sample_unit) }}</template>
              </el-table-column>
              <el-table-column label="累计开销" width="150">
                <template #default="{ row }">{{ formatSample(row.totalValue, currentProfile.sample_unit) }}</template>
              </el-table-column>
              <el-table-column label="自身占比" width="105">
                <template #default="{ row }">{{ row.percent.toFixed(2) }}%</template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import { getPprofRecord } from '@/api/pprof'

defineOptions({ name: 'PprofRecord' })

const route = useRoute()
const router = useRouter()
const chartElement = ref(null)
const loading = ref(false)
const errorMessage = ref('')
const selectedProfile = ref('cpu')
const viewTab = ref('flame')
const profiles = ref({})
const profileErrors = ref({})
const record = reactive({})
const statusLabels = { waiting: '等待中', collecting: '采集中', generating: '火焰图生成中', completed: '已完成', failed: '失败' }
const profileTypeOptions = [
  { label: 'CPU', value: 'cpu' },
  { label: 'Heap 内存', value: 'heap' },
  { label: 'Allocations', value: 'allocs' },
  { label: 'Goroutine', value: 'goroutine' },
  { label: 'Mutex', value: 'mutex' },
  { label: 'Block', value: 'block' },
  { label: 'Thread Create', value: 'threadcreate' },
  { label: 'Trace', value: 'trace' }
]
const interpretations = {
  cpu: {
    title: 'CPU：代码实际占用处理器的时间',
    description: '自身开销高的函数是直接消耗 CPU 的热点；累计开销高说明该函数及其下游调用整体较重。',
    tip: '优先检查 Top 10 中可优化的业务函数，并结合调用栈判断是否存在重复计算、序列化、正则或频繁查询。'
  },
  heap: {
    title: 'Heap 内存：完成 GC 后仍然存活的对象',
    description: '数值表示当前仍被引用、无法回收的内存，不等同于历史总分配量。持续增长可能意味着缓存过大或对象泄漏。',
    tip: '优先检查业务包中的大对象、无上限缓存、长期存活集合，以及没有释放引用的数据结构。'
  },
  allocs: {
    title: 'Allocations：进程运行以来的累计内存分配',
    description: '数值高不一定是泄漏，但表示该调用路径频繁创建对象，会增加 GC 压力并间接影响延迟。',
    tip: '重点关注循环内分配、字符串拼接、反射、JSON 编解码和可复用缓冲区。'
  },
  goroutine: {
    title: 'Goroutine：采样结束时各调用栈上的协程数量',
    description: '大量协程停在同一位置通常代表等待、堆积；数量长期上涨时需要警惕协程泄漏。',
    tip: '检查 channel 收发、锁等待、网络请求、定时任务及缺少超时或取消机制的调用。'
  },
  mutex: {
    title: 'Mutex：采样窗口内因互斥锁竞争产生的等待时间',
    description: '自身开销越高，说明该位置的锁竞争对并发性能影响越明显。没有数据通常表示采样期间竞争很少。',
    tip: '检查锁粒度、临界区耗时、热点共享状态，必要时拆分锁或缩短持锁时间。'
  },
  block: {
    title: 'Block：采样窗口内同步操作造成的阻塞等待时间',
    description: '主要反映 channel、select、条件变量等同步点上的等待；高值可能意味着生产消费不平衡或串行瓶颈。',
    tip: '检查无缓冲 channel、队列容量、上下游处理速度，以及可能永远等不到结果的同步逻辑。'
  },
  threadcreate: {
    title: 'Thread Create：触发操作系统线程创建的调用栈',
    description: '展示哪些运行路径促使 Go 运行时创建系统线程。数量异常高时可能与阻塞系统调用或线程锁定有关。',
    tip: '关注 syscall、CGO、runtime.LockOSThread，以及长时间阻塞系统线程的操作。'
  },
  trace: {
    title: 'Trace：调度、GC、系统调用和协程运行的时间线',
    description: 'Trace 不是按调用栈聚合的样本，不能准确转换为火焰图；原始时间线已随本次记录保存。',
    tip: '它适合定位调度延迟、GC 停顿、协程阻塞和并发时序。后续可接入专用时间线查看器。'
  }
}
let chart

const currentProfile = computed(() => profiles.value[selectedProfile.value] || {})
const currentInterpretation = computed(() => interpretations[selectedProfile.value] || interpretations.cpu)
const topOperations = computed(() => {
  const profile = currentProfile.value
  if (!profile?.root?.value) return []
  const rows = []
  const visit = (node, ancestors = []) => {
    const children = node.children || []
    const childTotal = children.reduce((total, child) => total + child.value, 0)
    const selfValue = Math.max(0, node.value - childTotal)
    const path = [...ancestors, node.name]
    if (node.name !== '全部' && selfValue > 0) {
      rows.push({
        name: node.name,
        path: path.filter((name) => name !== '全部').join(' → '),
        selfValue,
        totalValue: node.value,
        percent: selfValue * 100 / profile.root.value
      })
    }
    children.forEach((child) => visit(child, path))
  }
  visit(profile.root)
  return rows.sort((left, right) => right.selfValue - left.selfValue).slice(0, 10)
})

const loadRecord = async () => {
  loading.value = true
  errorMessage.value = ''
  try {
    const response = await getPprofRecord(route.params.id)
    Object.assign(record, response.data?.record || {})
    profiles.value = response.data?.profiles || (response.data?.profile ? { cpu: response.data.profile } : {})
    profileErrors.value = response.data?.profile_errors || {}
    await renderSelectedProfile()
  } finally {
    loading.value = false
  }
}

const renderSelectedProfile = async () => {
  chart?.dispose()
  chart = undefined
  const profile = profiles.value[selectedProfile.value]
  if (!profile?.root?.value) {
    errorMessage.value = profileErrors.value[selectedProfile.value] || (selectedProfile.value === 'trace'
      ? 'Trace 原始时间线已保存在本次记录目录。Trace 是时间线数据，不适合用火焰图展示。'
      : '该类型在本次采样中没有可展示的数据')
    return
  }
  errorMessage.value = ''
  if (viewTab.value !== 'flame') return
  await nextTick()
  renderFlameChart(profile)
}

const renderFlameChart = (profile) => {
  chart?.dispose()
  const root = profile.root
  const rows = []
  let maxDepth = 0
  const visit = (node, start, depth) => {
    let cursor = start
    maxDepth = Math.max(maxDepth, depth)
    for (const child of node.children || []) {
      const end = cursor + child.value
      rows.push([cursor, end, depth, child.name, child.value, root.value])
      visit(child, cursor, depth + 1)
      cursor = end
    }
  }
  visit(root, 0, 0)
  const minimumVisibleDepth = 20
  const axisMax = Math.max(maxDepth + 1, minimumVisibleDepth)
  chartElement.value.style.height = `${Math.max(660, (maxDepth + 4) * 25 + 110)}px`
  chart = echarts.init(chartElement.value)
  chart.setOption({
    animation: false,
    tooltip: { formatter: ({ value }) => `${value[3]}<br/>样本：${formatSample(value[4], profile.sample_unit)}<br/>占比：${(value[4] * 100 / value[5]).toFixed(2)}%` },
    dataZoom: [{ type: 'slider', xAxisIndex: 0, height: 20, bottom: 12, filterMode: 'none' }],
    grid: { left: 12, right: 12, top: 12, bottom: 92 },
    xAxis: { type: 'value', min: 0, max: root.value, show: false },
    yAxis: { type: 'value', min: -0.4, max: axisMax + 3, inverse: true, show: false },
    series: [{
      type: 'custom',
      encode: { x: [0, 1], y: 2 },
      data: rows,
      renderItem: (params, api) => {
        const depth = api.value(2)
        const start = api.coord([api.value(0), depth])
        const end = api.coord([api.value(1), depth + 1])
        const width = Math.max(0, end[0] - start[0] - 1)
        const height = Math.max(0, Math.abs(end[1] - start[1]) - 2)
        const name = api.value(3)
        return {
          type: 'rect',
          shape: { x: start[0], y: Math.min(start[1], end[1]) + 1, width, height },
          style: { fill: colorForName(name), stroke: 'rgba(255,255,255,.45)', lineWidth: 1 },
          textContent: width > 70 ? { type: 'text', style: { text: name, fill: '#fff', overflow: 'truncate', width: width - 8, fontSize: 11 } } : undefined,
          textConfig: { position: 'insideLeft', distance: 4 }
        }
      }
    }]
  })
}

const colorForName = (name) => {
  let hash = 0
  for (let index = 0; index < name.length; index += 1) hash = ((hash << 5) - hash + name.charCodeAt(index)) | 0
  return `hsl(${Math.abs(hash) % 360} 68% 58%)`
}
const formatDuration = (value = 0) => value >= 1e9 ? `${(value / 1e9).toFixed(2)} s` : value >= 1e6 ? `${(value / 1e6).toFixed(2)} ms` : `${(value / 1e3).toFixed(2)} μs`
const formatBytes = (value = 0) => {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1)
  return `${(value / (1024 ** index)).toFixed(index ? 1 : 0)} ${units[index]}`
}
const formatSample = (value, unit) => unit === 'bytes' ? formatBytes(value) : unit === 'count' ? `${value} 次` : formatDuration(value)
const formatSeconds = (value) => !value ? '即时快照' : value >= 60 ? `${value / 60} 分钟` : `${value} 秒`
const formatTime = (value) => value ? new Date(value).toLocaleString() : '-'
const resizeChart = () => chart?.resize()

onMounted(() => {
  loadRecord()
  window.addEventListener('resize', resizeChart)
})
onBeforeUnmount(() => {
  window.removeEventListener('resize', resizeChart)
  chart?.dispose()
})
</script>

<style scoped>
.pprof-detail-page { display: flex; flex-direction: column; gap: 16px; }
.detail-header, .chart-card { padding: 20px; }
.title-row { display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; }
.page-title { color: var(--el-text-color-primary); font-size: 18px; font-weight: 600; }
.page-tip { margin-top: 7px; color: var(--el-text-color-secondary); font-size: 13px; line-height: 1.6; }
.record-summary { margin-top: 18px; }
.chart-card { min-height: 650px; overflow-x: auto; }
.profile-tabs { margin-bottom: 12px; }
.view-tabs { margin-top: 8px; }
.interpretation-box { padding: 14px 16px; margin-bottom: 16px; border: 1px solid var(--el-border-color-lighter); border-radius: 6px; background: var(--el-fill-color-light); }
.interpretation-title { color: var(--el-text-color-primary); font-size: 15px; font-weight: 600; }
.interpretation-text, .interpretation-tip { margin-top: 7px; color: var(--el-text-color-regular); font-size: 13px; line-height: 1.6; }
.interpretation-tip { color: var(--el-text-color-secondary); }
.top-section { margin-bottom: 20px; }
.section-heading { display: flex; align-items: baseline; gap: 12px; margin-bottom: 10px; color: var(--el-text-color-primary); font-size: 15px; font-weight: 600; }
.section-heading small { color: var(--el-text-color-secondary); font-size: 12px; font-weight: 400; }
.flame-heading { margin-top: 6px; }
.top-table { width: 100%; }
.operation-name { color: var(--el-text-color-primary); font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; font-size: 12px; }
.operation-path { margin-top: 3px; overflow: hidden; color: var(--el-text-color-secondary); font-size: 11px; text-overflow: ellipsis; white-space: nowrap; }
.flame-chart { width: 100%; min-width: 900px; min-height: 600px; }
@media (max-width: 720px) { .title-row { align-items: stretch; flex-direction: column; } }
</style>
