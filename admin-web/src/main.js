import './style/element_visiable.scss'
import 'element-plus/theme-chalk/dark/css-vars.css'
import 'uno.css'
import { createApp } from 'vue'
import ElementPlus from 'element-plus'

import 'element-plus/dist/index.css'
// 引入 MeAdmin 前端初始化相关内容
import './core/meadmin'
// 引入封装的router
import router from '@/router/index'
import '@/permission'
import run from '@/core/meadmin.js'
import auth from '@/directive/auth'
import clickOutSide from '@/directive/clickOutSide'
import { store } from '@/pinia'
import App from './App.vue'
import '@/core/error-handel'

// 发布后旧入口页可能仍引用已替换的带 hash 资源。Vite 在预加载失败时会触发此事件，
// 自动刷新一次即可取得新入口；用 sessionStorage 防止网络故障时反复刷新。
window.addEventListener('vite:preloadError', (event) => {
  event.preventDefault()
  const reloadKey = 'vite-preload-error-reloaded'
  if (sessionStorage.getItem(reloadKey)) {
    return
  }
  sessionStorage.setItem(reloadKey, '1')
  window.location.reload()
})

const app = createApp(App)

app.config.productionTip = false

app
  .use(run)
  .use(ElementPlus)
  .use(store)
  .use(auth)
  .use(clickOutSide)
  .use(router)
  .mount('#app')
export default app
