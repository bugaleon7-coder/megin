<template>
  <div class="database-query-page">
    <div class="gva-card query-card">
      <div class="query-header">
        <div>
          <div class="query-title">数据库查询</div>
          <div class="query-tip">仅允许一条 SELECT 语句，并且必须包含最外层 LIMIT；禁止执行任何写入、结构变更及锁定操作。</div>
        </div>
      </div>
      <div class="table-picker">
        <span class="picker-label">数据表</span>
        <el-select
          v-model="selectedTable"
          filterable
          clearable
          :loading="tablesLoading"
          placeholder="请选择或搜索数据表"
          class="table-select"
          @change="handleTableChange"
        >
          <el-option v-for="table in tables" :key="table.name" :value="table.name" :label="table.name">
            <div class="table-option">
              <span>{{ table.name }}</span>
              <small>{{ table.comment || `约 ${table.rows} 行` }}</small>
            </div>
          </el-option>
        </el-select>
        <el-button type="primary" icon="search" :loading="loading" @click="executeQuery">执行查询</el-button>
        <el-button :disabled="!selectedTable" :loading="structureLoading" @click="showTableStructure">查看表结构</el-button>
      </div>
      <el-input
        v-model="sql"
        type="textarea"
        :rows="7"
        resize="vertical"
        spellcheck="false"
        placeholder="例如：SELECT id, username, created_at FROM sys_users ORDER BY id DESC LIMIT 20"
        @keydown="handleEditorKeydown"
      />
      <div class="query-shortcut">快捷键：Ctrl + Enter（macOS 为 Command + Enter）</div>
    </div>

    <div class="gva-table-box result-card">
      <div class="result-header">
        <span class="result-title">查询结果</span>
        <div v-if="executed" class="result-summary">
          <el-tag type="success" effect="plain">{{ result.row_count }} 行</el-tag>
          <span>耗时 {{ result.duration_ms }} ms</span>
          <el-tag v-if="result.truncated" type="warning" effect="plain">仅显示前 1000 行</el-tag>
        </div>
      </div>

      <el-empty v-if="!executed" description="输入查询语句后执行，字段和数据将在这里显示" />
      <el-empty v-else-if="result.columns.length === 0" description="查询未返回字段" />
      <el-table v-else :data="result.rows" border stripe height="520" class="query-result-table">
        <el-table-column type="index" label="#" width="64" fixed="left" />
        <el-table-column
          v-for="(column, index) in result.columns"
          :key="`${column.name}-${index}`"
          :label="column.name"
          min-width="160"
          show-overflow-tooltip
        >
          <template #header>
            <div class="column-header">
              <span>{{ column.name }}</span>
              <small v-if="column.database_type">{{ column.database_type }}</small>
            </div>
          </template>
          <template #default="{ row }">
            <span v-if="row[index] === null" class="null-value">NULL</span>
            <span v-else>{{ displayValue(row[index]) }}</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="structureVisible" :title="`${selectedTable} 表结构`" width="min(900px, 92vw)" destroy-on-close>
      <pre class="structure-sql">{{ createTableSQL }}</pre>
    </el-dialog>
  </div>
</template>

<script setup>
import { onMounted, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { executeDatabaseQuery, getDatabaseTables, getDatabaseTableStructure } from '@/api/databaseQuery'

defineOptions({ name: 'DatabaseQuery' })

const sql = ref('SELECT id, username, created_at\nFROM sys_users\nORDER BY id DESC\nLIMIT 20')
const loading = ref(false)
const executed = ref(false)
const tablesLoading = ref(false)
const structureLoading = ref(false)
const tables = ref([])
const selectedTable = ref('')
const structureVisible = ref(false)
const createTableSQL = ref('')
const result = reactive({ columns: [], rows: [], row_count: 0, duration_ms: 0, truncated: false })

const loadTables = async () => {
  tablesLoading.value = true
  try {
    const response = await getDatabaseTables()
    tables.value = response.data || []
  } finally {
    tablesLoading.value = false
  }
}

const handleTableChange = async (tableName) => {
  createTableSQL.value = ''
  if (!tableName) return
  const table = tables.value.find((item) => item.name === tableName)
  const escapedTable = tableName.replaceAll('`', '``')
  const escapedOrderColumn = table?.order_column?.replaceAll('`', '``')
  sql.value = escapedOrderColumn
    ? `SELECT *\nFROM \`${escapedTable}\`\nORDER BY \`${escapedOrderColumn}\` DESC\nLIMIT 10`
    : `SELECT *\nFROM \`${escapedTable}\`\nLIMIT 10`
  await executeQuery()
}

const showTableStructure = async () => {
  if (!selectedTable.value) return
  structureLoading.value = true
  try {
    const response = await getDatabaseTableStructure({ table_name: selectedTable.value })
    createTableSQL.value = response.data?.create_sql || ''
    structureVisible.value = true
  } finally {
    structureLoading.value = false
  }
}

const executeQuery = async () => {
  if (!sql.value.trim()) {
    ElMessage.warning('请输入 SQL')
    return
  }
  loading.value = true
  try {
    const response = await executeDatabaseQuery({ sql: sql.value })
    Object.assign(result, response.data || { columns: [], rows: [], row_count: 0, duration_ms: 0, truncated: false })
    executed.value = true
    ElMessage.success(`查询完成，共返回 ${result.row_count} 行`)
  } finally {
    loading.value = false
  }
}

const handleEditorKeydown = (event) => {
  if (event.key === 'Enter' && (event.ctrlKey || event.metaKey)) {
    event.preventDefault()
    executeQuery()
  }
}

const displayValue = (value) => {
  if (typeof value === 'object') return JSON.stringify(value)
  return String(value)
}

onMounted(loadTables)
</script>

<style scoped>
.database-query-page { display: flex; flex-direction: column; gap: 16px; }
.query-card { padding: 20px; }
.query-header { display: flex; align-items: flex-start; justify-content: space-between; gap: 24px; margin-bottom: 16px; }
.query-title, .result-title { color: var(--el-text-color-primary); font-size: 16px; font-weight: 600; }
.query-tip { margin-top: 7px; color: var(--el-text-color-secondary); font-size: 13px; line-height: 1.6; }
.table-picker { display: flex; align-items: center; gap: 12px; margin-bottom: 14px; }
.picker-label { flex: 0 0 auto; color: var(--el-text-color-regular); font-size: 14px; }
.table-select { width: min(460px, 100%); }
.table-option { display: flex; justify-content: space-between; gap: 24px; }
.table-option small { overflow: hidden; color: var(--el-text-color-placeholder); text-overflow: ellipsis; white-space: nowrap; }
.query-shortcut { margin-top: 8px; color: var(--el-text-color-placeholder); font-size: 12px; }
.result-card { min-height: 280px; }
.result-header { display: flex; align-items: center; justify-content: space-between; min-height: 32px; margin-bottom: 14px; }
.result-summary { display: flex; align-items: center; gap: 12px; color: var(--el-text-color-secondary); font-size: 13px; }
.query-result-table { width: 100%; }
.column-header { display: flex; flex-direction: column; line-height: 1.25; }
.column-header small { color: var(--el-text-color-placeholder); font-size: 10px; font-weight: 400; }
.null-value { color: var(--el-text-color-placeholder); font-style: italic; }
.structure-sql { max-height: 65vh; margin: 0; overflow: auto; padding: 16px; border: 1px solid var(--el-border-color); border-radius: 6px; background: var(--el-fill-color-light); color: var(--el-text-color-primary); font: 13px/1.65 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; white-space: pre-wrap; word-break: break-word; }

@media (max-width: 768px) {
  .query-header { align-items: stretch; flex-direction: column; gap: 12px; }
  .table-picker { align-items: stretch; flex-direction: column; }
  .table-select { width: 100%; }
  .result-header { align-items: flex-start; flex-direction: column; gap: 10px; }
  .result-summary { flex-wrap: wrap; }
}
</style>
