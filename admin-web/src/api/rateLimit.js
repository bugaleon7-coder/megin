import service from '@/utils/request'

// createRateLimitRule 新增全局或指定接口限流规则。
export const createRateLimitRule = (data) => {
  return service({
    url: '/system/rate-limit/create',
    method: 'post',
    data
  })
}

// updateRateLimitRule 修改限流规则，后端保存后会立即刷新当前实例。
export const updateRateLimitRule = (data) => {
  return service({
    url: '/system/rate-limit/update',
    method: 'put',
    data
  })
}

// changeRateLimitRuleStatus 启用或禁用限流规则。
export const changeRateLimitRuleStatus = (data) => {
  return service({
    url: '/system/rate-limit/changeStatus',
    method: 'put',
    data
  })
}

// deleteRateLimitRule 删除指定限流规则。
export const deleteRateLimitRule = (params) => {
  return service({
    url: '/system/rate-limit/delete',
    method: 'delete',
    params
  })
}

// getRateLimitRuleDetail 查询限流规则详情。
export const getRateLimitRuleDetail = (params) => {
  return service({
    url: '/system/rate-limit/detail',
    method: 'get',
    params
  })
}

// getRateLimitRulePage 分页查询限流规则。
export const getRateLimitRulePage = (params) => {
  return service({
    url: '/system/rate-limit/pageList',
    method: 'get',
    params
  })
}

// refreshRateLimitRules 从 MySQL 重新加载启用规则到当前服务实例内存。
export const refreshRateLimitRules = () => {
  return service({
    url: '/system/rate-limit/refresh',
    method: 'post'
  })
}
