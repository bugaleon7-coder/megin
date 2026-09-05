import service from '@/utils/request'

export const getScheduledTaskOptions = () => service({ url: '/system/scheduled-task/options', method: 'get' })
export const getScheduledTaskPage = (params) => service({ url: '/system/scheduled-task/pageList', method: 'get', params })
export const createScheduledTask = (data) => service({ url: '/system/scheduled-task/create', method: 'post', data })
export const updateScheduledTask = (data) => service({ url: '/system/scheduled-task/update', method: 'put', data })
export const deleteScheduledTask = (params) => service({ url: '/system/scheduled-task/delete', method: 'delete', params })
export const executeScheduledTask = (data) => service({ url: '/system/scheduled-task/execute', method: 'post', data })
export const getScheduledTaskLogs = (params) => service({ url: '/system/scheduled-task/logPageList', method: 'get', params })
