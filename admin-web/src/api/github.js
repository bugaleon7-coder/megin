import axios from 'axios'

const service = axios.create()

export function Commits(page) {
  return service({
    url:
      '/api/commits?page=' +
      page,
    method: 'get'
  })
}

export function Members() {
  return service({
    url: '/api/members',
    method: 'get'
  })
}
