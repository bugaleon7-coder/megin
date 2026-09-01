import { useUserStore } from '@/pinia/modules/user'
import { useRouterStore } from '@/pinia/modules/router'
import getPageTitle from '@/utils/page'
import { registerAsyncRoutes } from '@/utils/registerAsyncRoutes'
import router from '@/router'
import Nprogress from 'nprogress'
import 'nprogress/nprogress.css'

// 配置 NProgress
Nprogress.configure({
  showSpinner: false,
  ease: 'ease',
  speed: 500
})

// 白名单路由
const WHITE_LIST = ['Login', 'Init']

// 处理路由加载
const setupRouter = async (userStore) => {
  try {
    const routerStore = useRouterStore()
    await Promise.all([routerStore.SetAsyncRouter(), userStore.GetUserInfo()])
    registerAsyncRoutes(routerStore.asyncRouters || [])
    return true
  } catch (error) {
    console.error('Setup router failed:', error)
    return false
  }
}

// 路由守卫
router.beforeEach(async (to, from) => {
  const userStore = useUserStore()
  const routerStore = useRouterStore()
  const token = userStore.token

  Nprogress.start()

  // 处理元数据和缓存
  to.meta.matched = [...to.matched]
  await routerStore.handleKeepAlive(to)
  // 设置页面标题
  document.title = getPageTitle(to.meta.title, to)
  if (to.meta.client) {
    return true
  }

  // 白名单路由处理
  if (WHITE_LIST.includes(to.name)) {
    if (to.name === 'Login' && to.query.forceLogin === '1') {
      await userStore.ClearStorage()
      return true
    }
    if (token) {
      if(!routerStore.asyncRouterFlag){
        await setupRouter(userStore)
      }
      if(userStore.userInfo.authority.defaultRouter){
        return { name: userStore.userInfo.authority.defaultRouter }
      }
    }
    return  true
  }

  // 需要登录的路由处理
  if (token) {
    // 处理需要跳转到首页的情况
    if (sessionStorage.getItem('needToHome') === 'true') {
      sessionStorage.removeItem('needToHome')
      return { path: '/' }
    }

    // 处理异步路由
    if (!routerStore.asyncRouterFlag && !WHITE_LIST.includes(from.name)) {
      await setupRouter(userStore)
      return to
    }

    return to.matched.length ? true : { path: '/layout/404' }
  }

  // 未登录跳转登录页
  return {
    name: 'Login',
    query: {
      redirect: to.fullPath
    }
  }
})

// 路由加载完成
router.afterEach(() => {
  document.querySelector('.main-cont.main-right')?.scrollTo(0, 0)
  Nprogress.done()
})

// 路由错误处理
router.onError((error) => {
  console.error('Router error:', error)
  Nprogress.remove()
})
