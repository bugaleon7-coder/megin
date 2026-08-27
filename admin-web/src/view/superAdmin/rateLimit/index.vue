<template>
  <div>
    <div class="gva-card rate-limit-notice">
      <div class="notice-title">限流规则说明</div>
      <div class="notice-content">
        全局规则是所有 <code>/api</code> 接口的默认规则；指定接口存在同维度规则时，使用接口规则替代全局规则，不重复叠加。
        IP 规则适用于全部接口，UID 规则仅适用于已经通过 Token 认证并解析出用户 ID 的接口。
      </div>
    </div>

    <div class="gva-search-box">
      <el-form :inline="true" :model="searchInfo" @keyup.enter="onSubmit">
        <el-form-item label="规则名称">
          <el-input v-model="searchInfo.name" clearable placeholder="请输入规则名称" />
        </el-form-item>
        <el-form-item label="作用范围">
          <el-select v-model="searchInfo.scope_type" clearable placeholder="全部" style="width: 140px">
            <el-option label="全局规则" :value="1" />
            <el-option label="指定接口" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="限流维度">
          <el-select v-model="searchInfo.dimension" clearable placeholder="全部" style="width: 130px">
            <el-option label="IP" :value="1" />
            <el-option label="UID" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="searchInfo.status" clearable placeholder="全部" style="width: 120px">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="0" />
          </el-select>
        </el-form-item>
        <el-form-item label="请求方法">
          <el-select v-model="searchInfo.http_method" clearable placeholder="全部" style="width: 130px">
            <el-option v-for="method in methodOptions" :key="method" :label="method" :value="method" />
          </el-select>
        </el-form-item>
        <el-form-item label="接口路径">
          <el-input v-model="searchInfo.route_path" clearable placeholder="例如 /api/common/test" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" icon="search" @click="onSubmit">查询</el-button>
          <el-button icon="refresh" @click="onReset">重置</el-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="plus" @click="openCreateDialog">新增规则</el-button>
        <el-button icon="refresh" @click="handleRefreshRules">刷新内存规则</el-button>
      </div>

      <el-table :data="tableData" row-key="id" style="width: 100%">
        <el-table-column label="ID" prop="id" width="70" />
        <el-table-column label="规则名称" prop="name" min-width="150" show-overflow-tooltip />
        <el-table-column label="作用范围" width="110">
          <template #default="scope">
            <el-tag :type="scope.row.scope_type === 1 ? 'primary' : 'warning'">
              {{ scopeLabel(scope.row.scope_type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="接口" min-width="230" show-overflow-tooltip>
          <template #default="scope">
            <span v-if="scope.row.scope_type === 1" class="muted-text">全部 /api 接口</span>
            <span v-else><b>{{ scope.row.http_method }}</b> {{ scope.row.route_path }}</span>
          </template>
        </el-table-column>
        <el-table-column label="维度" width="90">
          <template #default="scope">
            <el-tag :type="scope.row.dimension === 1 ? 'success' : 'info'">
              {{ dimensionLabel(scope.row.dimension) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="平均速率" min-width="150">
          <template #default="scope">
            {{ formatRate(scope.row) }}
          </template>
        </el-table-column>
        <el-table-column label="桶容量" prop="burst" width="90" />
        <el-table-column label="状态" width="100">
          <template #default="scope">
            <el-switch
              v-model="scope.row.status"
              :active-value="1"
              :inactive-value="0"
              @change="(value) => handleStatusChange(scope.row, value)"
            />
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="175">
          <template #default="scope">{{ formatDate(scope.row.updated_at) }}</template>
        </el-table-column>
        <el-table-column label="备注" prop="remark" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" fixed="right" width="150">
          <template #default="scope">
            <el-button type="primary" link icon="edit" @click="openEditDialog(scope.row)">编辑</el-button>
            <el-button type="danger" link icon="delete" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <div class="gva-pagination">
        <el-pagination
          :current-page="pageNo"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <el-drawer
      v-model="dialogVisible"
      destroy-on-close
      :show-close="false"
      size="560px"
      :before-close="closeDialog"
    >
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg">{{ dialogType === 'create' ? '新增限流规则' : '编辑限流规则' }}</span>
          <div>
            <el-button @click="closeDialog">取消</el-button>
            <el-button type="primary" :loading="submitting" @click="submitDialog">确定</el-button>
          </div>
        </div>
      </template>

      <el-form ref="formRef" :model="formData" :rules="formRules" label-position="top">
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="formData.name" maxlength="100" show-word-limit placeholder="例如：全局 IP 限流" />
        </el-form-item>
        <el-form-item label="作用范围" prop="scope_type">
          <el-radio-group
            v-model="formData.scope_type"
            :disabled="dialogType === 'update'"
            @change="handleScopeChange"
          >
            <el-radio-button :value="1">全局规则</el-radio-button>
            <el-radio-button :value="2">指定接口</el-radio-button>
          </el-radio-group>
          <div class="form-help">
            <template v-if="dialogType === 'update'">
              编辑时不能改变规则作用范围；如需新增指定接口规则，请关闭窗口后点击“新增规则”。
            </template>
            <template v-else>
              同一个作用范围、接口和维度只能配置一条规则；接口规则优先于同维度全局规则。
            </template>
          </div>
        </el-form-item>
        <template v-if="formData.scope_type === 2">
          <el-form-item label="HTTP 请求方法" prop="http_method">
            <el-select
              v-model="formData.http_method"
              :disabled="dialogType === 'update'"
              placeholder="请选择请求方法"
              style="width: 100%"
            >
              <el-option v-for="method in methodOptions" :key="method" :label="method" :value="method" />
            </el-select>
          </el-form-item>
          <el-form-item label="Gin 路由模板" prop="route_path">
            <el-input
              v-model="formData.route_path"
              :disabled="dialogType === 'update'"
              maxlength="255"
              placeholder="例如 /api/article/:id"
            />
            <div class="form-help">必须填写以 /api/ 开头的后端 Gin 路由模板，不要填写域名和查询参数。</div>
          </el-form-item>
        </template>
        <el-form-item label="限流维度" prop="dimension">
          <el-radio-group v-model="formData.dimension" :disabled="dialogType === 'update'">
            <el-radio-button :value="1">按 IP</el-radio-button>
            <el-radio-button :value="2">按 UID</el-radio-button>
          </el-radio-group>
          <div v-if="formData.dimension === 2" class="form-help">
            UID 来自 Token 解析结果，只对需要登录的 /api 接口生效。
          </div>
        </el-form-item>

        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="每周期令牌数" prop="rate_count">
              <el-input-number v-model="formData.rate_count" :min="1" :max="1000000" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="周期（秒）" prop="interval_seconds">
              <el-input-number v-model="formData.interval_seconds" :min="1" :max="86400" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <div class="rate-preview">
          当前平均速率：每 {{ formData.interval_seconds }} 秒补充 {{ formData.rate_count }} 个令牌
          （约 {{ calculatedRate }} 个/秒）
        </div>

        <el-form-item label="桶容量（Burst）" prop="burst">
          <el-input-number v-model="formData.burst" :min="1" :max="1000000" controls-position="right" style="width: 100%" />
          <div class="form-help">决定空闲后最多可积攒的令牌数，也是一次瞬时请求最多可放行的数量。</div>
        </el-form-item>
        <el-form-item label="启用状态" prop="status">
          <el-switch v-model="formData.status" :active-value="1" :inactive-value="0" active-text="启用" inactive-text="禁用" />
        </el-form-item>
        <el-form-item label="备注" prop="remark">
          <el-input v-model="formData.remark" type="textarea" :rows="3" maxlength="500" show-word-limit placeholder="请输入规则用途或调整原因" />
        </el-form-item>
      </el-form>
    </el-drawer>
  </div>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { formatDate } from '@/utils/format'
import {
  changeRateLimitRuleStatus,
  createRateLimitRule,
  deleteRateLimitRule,
  getRateLimitRuleDetail,
  getRateLimitRulePage,
  refreshRateLimitRules,
  updateRateLimitRule
} from '@/api/rateLimit'

defineOptions({
  name: 'ApiRateLimit'
})

const methodOptions = ['GET', 'POST', 'PUT', 'DELETE', 'PATCH']
const pageNo = ref(1)
const pageSize = ref(20)
const total = ref(0)
const tableData = ref([])
const dialogVisible = ref(false)
const dialogType = ref('create')
const submitting = ref(false)
const formRef = ref(null)

const searchInfo = reactive({
  name: '',
  scope_type: undefined,
  dimension: undefined,
  status: undefined,
  http_method: '',
  route_path: ''
})

const defaultFormData = () => ({
  id: undefined,
  name: '',
  scope_type: 1,
  http_method: '*',
  route_path: '*',
  dimension: 1,
  rate_count: 1,
  interval_seconds: 1,
  burst: 1,
  status: 1,
  remark: ''
})

const formData = reactive(defaultFormData())

const validateRoutePath = (_rule, value, callback) => {
  if (formData.scope_type !== 2) {
    callback()
    return
  }
  if (!value) {
    callback(new Error('请输入 Gin 路由模板'))
    return
  }
  if (!value.startsWith('/api/')) {
    callback(new Error('接口路径必须以 /api/ 开头'))
    return
  }
  callback()
}

const formRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  scope_type: [{ required: true, message: '请选择作用范围', trigger: 'change' }],
  http_method: [{ required: true, message: '请选择 HTTP 请求方法', trigger: 'change' }],
  route_path: [{ validator: validateRoutePath, trigger: 'blur' }],
  dimension: [{ required: true, message: '请选择限流维度', trigger: 'change' }],
  rate_count: [{ required: true, message: '请输入每周期令牌数', trigger: 'change' }],
  interval_seconds: [{ required: true, message: '请输入周期秒数', trigger: 'change' }],
  burst: [{ required: true, message: '请输入桶容量', trigger: 'change' }]
}

const calculatedRate = computed(() => {
  const interval = Number(formData.interval_seconds) || 1
  const count = Number(formData.rate_count) || 0
  return (count / interval).toFixed(4).replace(/0+$/, '').replace(/\.$/, '')
})

const scopeLabel = (value) => (value === 1 ? '全局规则' : '指定接口')
const dimensionLabel = (value) => (value === 1 ? 'IP' : 'UID')
const formatRate = (row) => `每 ${row.interval_seconds} 秒 ${row.rate_count} 个`

const resetFormData = (data = {}) => {
  Object.assign(formData, defaultFormData(), data)
}

const getTableData = async () => {
  const params = {
    page_no: pageNo.value,
    page_size: pageSize.value,
    ...searchInfo
  }
  Object.keys(params).forEach((key) => {
    if (params[key] === '' || params[key] === undefined || params[key] === null) {
      delete params[key]
    }
  })

  const response = await getRateLimitRulePage(params)
  if (response.code === 0) {
    tableData.value = response.data.list || []
    total.value = response.data.total_size || 0
    pageNo.value = response.data.page_no || 1
    pageSize.value = response.data.page_size || pageSize.value
  }
}

const onSubmit = () => {
  pageNo.value = 1
  getTableData()
}

const onReset = () => {
  Object.assign(searchInfo, {
    name: '',
    scope_type: undefined,
    dimension: undefined,
    status: undefined,
    http_method: '',
    route_path: ''
  })
  pageNo.value = 1
  getTableData()
}

const handleCurrentChange = (value) => {
  pageNo.value = value
  getTableData()
}

const handleSizeChange = (value) => {
  pageSize.value = value
  pageNo.value = 1
  getTableData()
}

const handleScopeChange = (value) => {
  if (value === 1) {
    formData.http_method = '*'
    formData.route_path = '*'
  } else {
    formData.http_method = 'GET'
    formData.route_path = '/api/'
  }
  formRef.value?.clearValidate(['http_method', 'route_path'])
}

const openCreateDialog = () => {
  dialogType.value = 'create'
  resetFormData()
  dialogVisible.value = true
}

const openEditDialog = async (row) => {
  const response = await getRateLimitRuleDetail({ id: row.id })
  if (response.code !== 0) return

  dialogType.value = 'update'
  resetFormData(response.data)
  dialogVisible.value = true
}

const closeDialog = () => {
  dialogVisible.value = false
  formRef.value?.resetFields()
}

const submitDialog = async () => {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  const payload = { ...formData }
  if (payload.scope_type === 1) {
    payload.http_method = '*'
    payload.route_path = '*'
  }

  submitting.value = true
  try {
    const response = dialogType.value === 'create'
      ? await createRateLimitRule(payload)
      : await updateRateLimitRule(payload)
    if (response.code === 0) {
      ElMessage.success(dialogType.value === 'create' ? '新增成功' : '修改成功')
      closeDialog()
      getTableData()
    }
  } finally {
    submitting.value = false
  }
}

const handleStatusChange = async (row, value) => {
  const previousStatus = value === 1 ? 0 : 1
  try {
    const response = await changeRateLimitRuleStatus({ id: row.id, status: value })
    if (response.code === 0) {
      ElMessage.success(value === 1 ? '规则已启用' : '规则已禁用')
      return
    }
  } catch (_error) {
    // 请求失败时由统一请求层展示错误，这里只负责恢复开关状态。
  }
  row.status = previousStatus
}

const handleDelete = async (row) => {
  const confirmed = await ElMessageBox.confirm(
    `确定删除限流规则“${row.name}”吗？删除后对应接口将回退使用同维度全局规则。`,
    '删除确认',
    { type: 'warning', confirmButtonText: '确定', cancelButtonText: '取消' }
  ).then(() => true).catch(() => false)
  if (!confirmed) return

  const response = await deleteRateLimitRule({ id: row.id })
  if (response.code === 0) {
    ElMessage.success('删除成功')
    if (tableData.value.length === 1 && pageNo.value > 1) pageNo.value--
    getTableData()
  }
}

const handleRefreshRules = async () => {
  const response = await refreshRateLimitRules()
  if (response.code === 0) {
    ElMessage.success(`已刷新 ${response.data.rule_count} 条启用规则`)
    getTableData()
  }
}

getTableData()
</script>

<style scoped>
.rate-limit-notice {
  margin-bottom: 12px;
  padding: 16px 20px;
}

.notice-title {
  margin-bottom: 6px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.notice-content,
.form-help,
.muted-text {
  color: var(--el-text-color-secondary);
}

.notice-content {
  line-height: 1.7;
}

.notice-content code {
  padding: 1px 5px;
  border-radius: 4px;
  background: var(--el-fill-color-light);
}

.form-help {
  width: 100%;
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.5;
}

.rate-preview {
  margin: -4px 0 18px;
  padding: 10px 12px;
  border-radius: 6px;
  color: var(--el-color-primary);
  background: var(--el-color-primary-light-9);
}
</style>
