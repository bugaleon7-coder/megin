import service from '@/utils/request'

export const getPprofStatus = () => service({
  url: '/system/pprof/status',
  method: 'get'
})

export const togglePprof = (enable) => service({
  url: '/system/pprof/toggle',
  method: 'post',
  data: { enable }
})

export const startPprofCPUProfile = (durationSeconds) => service({
  url: '/system/pprof/cpu/start',
  method: 'post',
  data: { duration_seconds: durationSeconds }
})

export const getPprofCPUProfile = () => service({
  url: '/system/pprof/cpu/result',
  method: 'get'
})

export const getPprofRecords = (params) => service({
  url: '/system/pprof/records',
  method: 'get',
  params
})

export const getPprofRecord = (id) => service({
  url: '/system/pprof/record',
  method: 'get',
  params: { id }
})
