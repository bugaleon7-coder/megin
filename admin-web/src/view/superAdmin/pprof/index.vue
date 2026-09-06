<template>
  <div class="pprof-page">
    <div class="gva-card control-card">
      <div class="control-main">
        <div>
          <div class="page-title">Pprof 性能分析</div>
          <div class="page-tip">每次采样生成一条综合记录，同时收集 CPU、内存、协程、锁、阻塞、线程创建与 Trace 数据。</div>
        </div>
        <div class="switch-box">
          <span>{{ status.enabled ? '已开启' : '已关闭' }}</span>
          <el-switch v-model="switchValue" :loading="toggleLoading" @change="handleToggle" />
        </div>
      </div>
      <el-alert v-if="status.last_error" :title="status.last_error" type="error" :closable="false" show-icon />
      <div class="status-line">
        <el-tag :type="status.enabled ? 'success' : 'info'" effect="plain">{{ status.enabled ? '运行中' : '未运行' }}</el-tag>
        <span>监听地址：{{ status.address || '-' }}</span>
        <span>仅服务器本机可直接访问</span>
      </div>
    </div>

    <div class="metrics-grid">
      <div class="gva-card metric-card"><span>Goroutine</span><strong>{{ status.goroutines }}</strong></div>
      <div class="gva-card metric-card"><span>Heap Alloc</span><strong>{{ formatBytes(status.heap_alloc) }}</strong></div>
      <div class="gva-card metric-card"><span>Heap In Use</span><strong>{{ formatBytes(status.heap_in_use) }}</strong></div>
      <div class="gva-card metric-card"><span>Stack In Use</span><strong>{{ formatBytes(status.stack_in_use) }}</strong></div>
      <div class="gva-card metric-card"><span>GC 次数</span><strong>{{ status.gc_count }}</strong></div>
      <div class="gva-card metric-card"><span>最近 GC 停顿</span><strong>{{ formatDuration(status.last_gc_pause_ns) }}</strong></div>
    </div>

    <div class="gva-card records-card">
      <div class="records-header">
        <div>
          <div class="section-title">采样记录</div>
          <div class="section-tip">任务按创建顺序排队执行；进入已完成记录后可切换查看各类分析结果。</div>
        </div>
        <div class="profile-actions">
          <el-select v-model="durationSeconds" class="duration-select">
            <el-option v-for="item in durationOptions" :key="item.value" :label="item.label" :value="item.value" />
          </el-select>
          <el-button type="primary" :disabled="!status.enabled" :loading="creating" @click="startProfile">新建采样</el-button>
          <el-button class="refresh-button" :loading="loading" @click="loadRecords()">刷新</el-button>
        </div>
      </div>

      <el-table :data="records" border stripe v-loading="loading" :span-method="recordSpanMethod" class="records-table">
        <el-table-column prop="id" label="ID" width="76" />
        <el-table-column label="状态" width="120">
          <template #default="{ row }">
            <el-tag :type="statusMeta[row.status]?.type || 'info'" effect="plain">{{ statusMeta[row.status]?.label || row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="采样时长" width="110">
          <template #default="{ row }">
            <div v-if="isActiveRecord(row)" class="record-progress">
              <el-progress
                :percentage="recordProgress(row)"
                :status="row.status === 'generating' ? 'success' : undefined"
                :stroke-width="14"
                :striped="row.status === 'collecting'"
                striped-flow
              >
                <span class="progress-text">{{ recordProgressText(row) }}</span>
              </el-progress>
            </div>
            <span v-else>{{ formatSeconds(row.duration_seconds) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="CPU 样本" min-width="110">
          <template #default="{ row }">{{ row.sample_total ? formatSample(row.sample_total, row.sample_unit) : '-' }}</template>
        </el-table-column>
        <el-table-column label="创建时间" min-width="170">
          <template #default="{ row }">{{ formatTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="开始时间" min-width="170">
          <template #default="{ row }">{{ formatTime(row.started_at) }}</template>
        </el-table-column>
        <el-table-column label="完成时间" min-width="170">
          <template #default="{ row }">{{ formatTime(row.finished_at) }}</template>
        </el-table-column>
        <el-table-column prop="error_message" label="错误信息" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link :disabled="row.status !== 'completed'" @click="showRecord(row)">查看分析</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-box">
        <el-pagination
          v-model:current-page="page.page_no"
          v-model:page-size="page.page_size"
          :total="page.total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @current-change="loadRecords()"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

  </div>
</template>

<script setup>
import { onBeforeUnmount, onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useRouter } from 'vue-router'
import { getPprofRecords, getPprofStatus, startPprofCPUProfile, togglePprof } from '@/api/pprof'

defineOptions({ name: 'PprofMonitor' })

const durationOptions = [
  { label: '30 秒（推荐）', value: 30 },
  { label: '1 分钟', value: 60 },
  { label: '5 分钟', value: 300 },
  { label: '10 分钟', value: 600 }
]
const statusMeta = {
  waiting: { label: '等待中', type: 'info' },
  collecting: { label: '采集中', type: 'warning' },
  generating: { label: '火焰图生成中', type: 'warning' },
  completed: { label: '已完成', type: 'success' },
  failed: { label: '失败', type: 'danger' }
}
const status = reactive({ enabled: false, address: '', last_error: '', goroutines: 0, heap_alloc: 0, heap_in_use: 0, stack_in_use: 0, gc_count: 0, last_gc_pause_ns: 0 })
const page = reactive({ page_no: 1, page_size: 10, total: 0 })
const records = ref([])
const switchValue = ref(false)
const toggleLoading = ref(false)
const loading = ref(false)
const creating = ref(false)
const durationSeconds = ref(30)
const router = useRouter()
let pollTimer

const loadStatus = async () => {
  const response = await getPprofStatus()
  Object.assign(status, response.data || {})
  switchValue.value = Boolean(status.enabled)
}

const loadRecords = async (silent = false) => {
  if (!silent) loading.value = true
  try {
    const response = await getPprofRecords({ page_no: page.page_no, page_size: page.page_size })
    records.value = response.data?.list || []
    page.total = response.data?.total_size || 0
    updatePolling()
  } finally {
    if (!silent) loading.value = false
  }
}

const handleSizeChange = () => {
  page.page_no = 1
  loadRecords()
}

const handleToggle = async (enable) => {
  toggleLoading.value = true
  try {
    const response = await togglePprof(enable)
    Object.assign(status, response.data || {})
    switchValue.value = Boolean(status.enabled)
    ElMessage.success(enable ? 'Pprof 已开启' : 'Pprof 已关闭')
  } catch (_) {
    switchValue.value = status.enabled
  } finally {
    toggleLoading.value = false
  }
}

const startProfile = async () => {
  creating.value = true
  try {
    await startPprofCPUProfile(durationSeconds.value)
    page.page_no = 1
    await loadRecords()
    ElMessage.success('采样记录已创建')
  } finally {
    creating.value = false
  }
}

const updatePolling = () => {
  const active = records.value.some((item) => ['waiting', 'collecting', 'generating'].includes(item.status))
  if (active && !pollTimer) {
    pollTimer = window.setInterval(async () => {
      try {
        await Promise.all([loadRecords(true), loadStatus()])
      } catch (_) {
        // 请求层统一提示，下一轮继续恢复。
      }
    }, 1500)
  } else if (!active && pollTimer) {
    window.clearInterval(pollTimer)
    pollTimer = undefined
  }
}

const showRecord = async (record) => {
  await router.push({ name: 'pprofRecord', params: { id: record.id } })
}
const isActiveRecord = (record) => ['waiting', 'collecting', 'generating'].includes(record.status)
const recordProgress = (record) => {
  if (record.status === 'waiting') return 2
  if (record.status === 'generating') return 96
  const startedAt = record.started_at ? new Date(record.started_at).getTime() : Date.now()
  const elapsedSeconds = Math.max(0, (Date.now() - startedAt) / 1000)
  return Math.min(92, Math.max(5, Math.round(elapsedSeconds * 92 / record.duration_seconds)))
}
const recordProgressText = (record) => {
  if (record.status === 'waiting') return '等待采样任务执行'
  if (record.status === 'generating') return '采样完成，正在生成分析文件'
  const startedAt = record.started_at ? new Date(record.started_at).getTime() : Date.now()
  const elapsedSeconds = Math.min(record.duration_seconds, Math.max(0, Math.floor((Date.now() - startedAt) / 1000)))
  return `正在采集 ${elapsedSeconds} / ${record.duration_seconds} 秒`
}
const recordSpanMethod = ({ row, columnIndex }) => {
  if (!isActiveRecord(row)) return [1, 1]
  if (columnIndex === 2) return [1, 6]
  if (columnIndex >= 3 && columnIndex <= 7) return [0, 0]
  return [1, 1]
}
const formatBytes = (value = 0) => {
  if (!value) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB']
  const index = Math.min(Math.floor(Math.log(value) / Math.log(1024)), units.length - 1)
  return `${(value / (1024 ** index)).toFixed(index ? 1 : 0)} ${units[index]}`
}
const formatDuration = (value = 0) => value >= 1e9 ? `${(value / 1e9).toFixed(2)} s` : value >= 1e6 ? `${(value / 1e6).toFixed(2)} ms` : `${(value / 1e3).toFixed(2)} μs`
const formatSample = (value, unit) => unit === 'bytes' ? formatBytes(value) : unit === 'count' ? `${value} 次` : formatDuration(value)
const formatSeconds = (value) => !value ? '即时快照' : value >= 60 ? `${value / 60} 分钟` : `${value} 秒`
const formatTime = (value) => value ? new Date(value).toLocaleString() : '-'

onMounted(async () => {
  await Promise.all([loadStatus(), loadRecords()])
})
onBeforeUnmount(() => {
  if (pollTimer) window.clearInterval(pollTimer)
})
</script>

<style scoped>
.pprof-page { display: flex; flex-direction: column; gap: 16px; }
.control-card, .records-card { padding: 20px; }
.control-main, .records-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 20px; }
.page-title, .section-title { color: var(--el-text-color-primary); font-size: 16px; font-weight: 600; }
.page-tip, .section-tip { margin-top: 7px; color: var(--el-text-color-secondary); font-size: 13px; line-height: 1.6; }
.switch-box, .status-line, .profile-actions { display: flex; align-items: center; gap: 12px; }
.switch-box { flex: 0 0 auto; color: var(--el-text-color-regular); font-size: 14px; }
.status-line { flex-wrap: wrap; margin-top: 16px; color: var(--el-text-color-secondary); font-size: 13px; }
.control-card :deep(.el-alert) { margin-top: 16px; }
.metrics-grid { display: grid; grid-template-columns: repeat(6, minmax(0, 1fr)); gap: 12px; }
.metric-card { display: flex; min-width: 0; padding: 16px; flex-direction: column; gap: 9px; }
.metric-card span { color: var(--el-text-color-secondary); font-size: 12px; }
.metric-card strong { overflow: hidden; color: var(--el-text-color-primary); font-size: 20px; text-overflow: ellipsis; white-space: nowrap; }
.duration-select { width: 150px; }
.refresh-button { width: 68px; }
.records-table { width: 100%; margin-top: 18px; }
.record-progress { padding: 0 18px; }
.record-progress :deep(.el-progress-bar__outer) { background: var(--el-fill-color-darker); }
.progress-text { color: var(--el-text-color-primary); font-size: 12px; white-space: nowrap; }
.pagination-box { display: flex; justify-content: flex-end; margin-top: 16px; }
@media (max-width: 1100px) { .metrics-grid { grid-template-columns: repeat(3, minmax(0, 1fr)); } }
@media (max-width: 720px) {
  .control-main, .records-header { align-items: stretch; flex-direction: column; }
  .profile-actions { align-items: stretch; flex-direction: column; }
  .duration-select { width: 100%; }
  .metrics-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
}
</style>
