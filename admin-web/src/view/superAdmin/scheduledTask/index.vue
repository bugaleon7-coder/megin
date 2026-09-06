<template>
  <div>
    <div class="gva-card task-notice">
      <div class="notice-title">定时任务说明</div>
      <div class="notice-content">可配置 Cron 任务、手动执行任务，并查看每次执行的输出日志。<el-button type="primary" link @click="cronHelpVisible = true">查看 Cron 使用说明</el-button></div>
    </div>

    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo" @keyup.enter="loadTasks">
        <el-form-item label="任务名称"><el-input v-model="searchInfo.name" clearable placeholder="请输入任务名称" /></el-form-item>
        <el-form-item label="状态"><el-select v-model="searchInfo.status" clearable placeholder="全部" style="width: 120px"><el-option label="启用" :value="1" /><el-option label="禁用" :value="0" /></el-select></el-form-item>
        <el-form-item><el-button type="primary" icon="search" @click="loadTasks">查询</el-button><el-button icon="refresh" @click="resetSearch">重置</el-button></el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list"><el-button type="primary" icon="plus" @click="openCreate">新增任务</el-button></div>
      <el-table :data="tableData" row-key="id">
        <el-table-column prop="id" label="ID" width="70" />
        <el-table-column prop="name" label="任务名称" min-width="160" />
        <el-table-column label="执行器" min-width="130" show-overflow-tooltip><template #default="{ row }">{{ jobName(row.job_key) }}</template></el-table-column>
        <el-table-column prop="cron_expr" label="Cron 表达式" min-width="120"><template #default="{ row }"><code>{{ row.cron_expr }}</code></template></el-table-column>
        <el-table-column label="状态" width="145"><template #default="{ row }"><el-switch :model-value="row.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" @change="(value) => changeStatus(row, value)" /></template></el-table-column>
        <el-table-column prop="remark" label="备注" min-width="160" show-overflow-tooltip />
        <el-table-column label="更新时间" width="175"><template #default="{ row }">{{ formatDate(row.updated_at) }}</template></el-table-column>
        <el-table-column label="操作" fixed="right" width="280"><template #default="{ row }"><el-button type="primary" link icon="video-play" @click="execute(row)">执行</el-button><el-button type="primary" link icon="document" @click="openLogs(row)">日志</el-button><el-button type="primary" link icon="edit" @click="openEdit(row)">编辑</el-button><el-button type="danger" link icon="delete" @click="remove(row)">删除</el-button></template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination :current-page="pageNo" :page-size="pageSize" :total="total" layout="total, prev, pager, next" @current-change="changePage" /></div>
    </div>

    <el-drawer v-model="formVisible" size="520px" :show-close="false" destroy-on-close>
      <template #header><div class="flex justify-between items-center"><span class="text-lg">{{ editing ? '编辑定时任务' : '新增定时任务' }}</span><div><el-button @click="formVisible = false">取消</el-button><el-button type="primary" :loading="submitting" @click="submit">确定</el-button></div></div></template>
      <el-form ref="formRef" :model="formData" :rules="rules" label-position="top">
        <el-form-item label="任务名称" prop="name"><el-input v-model="formData.name" maxlength="100" placeholder="例如：每日限流规则刷新" /></el-form-item>
        <el-form-item label="任务执行器" prop="job_key"><el-select v-model="formData.job_key" placeholder="请选择执行器" style="width: 100%"><el-option v-for="item in options" :key="item.key" :label="item.name" :value="item.key" /></el-select></el-form-item>
        <el-form-item label="Cron 表达式" prop="cron_expr"><el-input v-model="formData.cron_expr" placeholder="例如：0 2 * * *" /><div class="form-help">依次为：分、时、日、月、周；支持 <code>@daily</code> 等描述符。</div></el-form-item>
        <el-form-item label="状态"><el-switch v-model="formData.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="formData.remark" type="textarea" :rows="3" maxlength="500" show-word-limit /></el-form-item>
      </el-form>
    </el-drawer>

    <el-drawer v-model="logVisible" size="760px" :title="`${currentTask?.name || ''} - 执行日志`" destroy-on-close>
      <el-table :data="logs" border>
        <el-table-column label="触发方式" width="90"><template #default="{ row }">{{ row.trigger_type === 'manual' ? '手动' : 'Cron' }}</template></el-table-column>
        <el-table-column label="结果" width="80"><template #default="{ row }"><el-tag :type="logStatusType(row.status)">{{ logStatusLabel(row.status) }}</el-tag></template></el-table-column>
        <el-table-column prop="output" label="输出" min-width="300"><template #default="{ row }"><pre class="task-output">{{ row.output }}</pre></template></el-table-column>
        <el-table-column label="开始时间" width="175"><template #default="{ row }">{{ formatDate(row.started_at) }}</template></el-table-column>
        <el-table-column label="耗时" width="80"><template #default="{ row }">{{ row.duration_ms }} ms</template></el-table-column>
      </el-table>
      <div class="gva-pagination"><el-pagination :current-page="logPageNo" :page-size="20" :total="logTotal" layout="total, prev, pager, next" @current-change="loadLogs" /></div>
    </el-drawer>

    <el-dialog v-model="cronHelpVisible" title="Cron 使用说明" width="520px">
      <div class="cron-help">
        <p>Cron 使用标准五段格式：分 时 日 月 周。</p>
        <p>任务只能选择服务端注册的执行器；每次自动或手动执行都会保存输出日志。</p>
        <div class="cron-example"><code>0 * * * *</code><span>每小时整点</span></div>
        <div class="cron-example"><code>0 2 * * *</code><span>每天 02:00</span></div>
        <div class="cron-example"><code>*/15 * * * *</code><span>每 15 分钟</span></div>
        <div class="cron-example"><code>0 9 * * 1-5</code><span>工作日 09:00</span></div>
        <div class="cron-example"><code>0 8 * * 1</code><span>每周一 08:00</span></div>
        <div class="cron-example"><code>0 0 1 * *</code><span>每月 1 日 00:00</span></div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { reactive, ref, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDate } from '@/utils/format'
import { createScheduledTask, deleteScheduledTask, executeScheduledTask, getScheduledTaskLogs, getScheduledTaskOptions, getScheduledTaskPage, updateScheduledTask } from '@/api/scheduledTask'

defineOptions({ name: 'ScheduledTask' })
const options = ref([]); const tableData = ref([]); const pageNo = ref(1); const pageSize = ref(20); const total = ref(0)
const searchInfo = reactive({ name: '', status: undefined }); const formVisible = ref(false); const formRef = ref(); const submitting = ref(false); const editing = ref(false)
const formData = reactive({ id: undefined, name: '', job_key: '', cron_expr: '', status: 1, remark: '' })
const rules = { name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }], job_key: [{ required: true, message: '请选择任务执行器', trigger: 'change' }], cron_expr: [{ required: true, message: '请输入 Cron 表达式', trigger: 'blur' }] }
const cronHelpVisible = ref(false)
const logVisible = ref(false); const logs = ref([]); const currentTask = ref(null); const logPageNo = ref(1); const logTotal = ref(0)
const jobName = (key) => options.value.find((item) => item.key === key)?.name || key
const logStatusLabel = (status) => ({ 1: '成功', 2: '失败', 3: '跳过' })[status] || '未知'
const logStatusType = (status) => ({ 1: 'success', 2: 'danger', 3: 'warning' })[status] || 'info'
const resetForm = () => Object.assign(formData, { id: undefined, name: '', job_key: '', cron_expr: '', status: 1, remark: '' })
const loadTasks = async () => { const res = await getScheduledTaskPage({ page_no: pageNo.value, page_size: pageSize.value, ...searchInfo }); tableData.value = res.data.list; total.value = res.data.total_size }
const resetSearch = () => { searchInfo.name = ''; searchInfo.status = undefined; pageNo.value = 1; loadTasks() }
const changePage = (page) => { pageNo.value = page; loadTasks() }
const openCreate = () => { editing.value = false; resetForm(); formVisible.value = true }
const openEdit = (row) => { editing.value = true; Object.assign(formData, row); formVisible.value = true }
const submit = async () => { await formRef.value.validate(); submitting.value = true; try { if (editing.value) await updateScheduledTask(formData); else await createScheduledTask(formData); ElMessage.success('保存成功'); formVisible.value = false; loadTasks() } finally { submitting.value = false } }
const changeStatus = async (row, nextStatus) => { const action = nextStatus === 1 ? '启用' : '禁用'; try { await ElMessageBox.confirm(`确认${action}任务“${row.name}”吗？`, `${action}确认`, { type: 'warning' }); await updateScheduledTask({ ...row, status: nextStatus }); ElMessage.success(`任务已${action}`) } catch (error) { if (error !== 'cancel' && error !== 'close') ElMessage.error('状态更新失败') } finally { loadTasks() } }
const remove = async (row) => { await ElMessageBox.confirm(`确认删除任务“${row.name}”吗？`, '删除确认', { type: 'warning' }); await deleteScheduledTask({ id: row.id }); ElMessage.success('删除成功'); loadTasks() }
const execute = async (row) => { await ElMessageBox.confirm(`确认立即执行任务“${row.name}”吗？`, '手动执行', { type: 'warning' }); const res = await executeScheduledTask({ id: row.id }); ElMessage.success(res.data.status === 1 ? '执行成功，已生成日志' : '执行失败，已生成日志'); openLogs(row) }
const openLogs = (row) => { currentTask.value = row; logPageNo.value = 1; logVisible.value = true; loadLogs() }
const loadLogs = async (page) => { if (page) logPageNo.value = page; const res = await getScheduledTaskLogs({ task_id: currentTask.value.id, page_no: logPageNo.value, page_size: 20 }); logs.value = res.data.list; logTotal.value = res.data.total_size }
onMounted(async () => { const res = await getScheduledTaskOptions(); options.value = res.data; loadTasks() })
</script>

<style scoped>
.task-notice { margin-bottom: 16px; padding: 16px 20px; }
.notice-title { font-size: 16px; font-weight: 600; margin-bottom: 8px; }
.notice-content, .form-help { color: var(--el-text-color-secondary); line-height: 1.7; font-size: 13px; }
.cron-help p { margin: 0 0 10px; line-height: 1.7; }
.cron-example { display: flex; align-items: center; gap: 16px; padding: 8px 0; border-bottom: 1px solid var(--el-border-color-lighter); }
.cron-example code { min-width: 130px; }
.task-output { margin: 0; white-space: pre-wrap; font-family: inherit; }
</style>
