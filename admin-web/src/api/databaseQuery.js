import service from '@/utils/request'

export const getDatabaseTables = () => service({
  url: '/system/database-query/tables',
  method: 'get'
})

export const getDatabaseTableStructure = (params) => service({
  url: '/system/database-query/table-structure',
  method: 'get',
  params
})

// executeDatabaseQuery 执行受限的只读 SELECT 查询。
export const executeDatabaseQuery = (data) => service({
  url: '/system/database-query/execute',
  method: 'post',
  data
})
