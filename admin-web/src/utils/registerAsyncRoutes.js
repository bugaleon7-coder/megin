import router from '@/router'

function isExternalUrl(val) {
  return typeof val === 'string' && /^(https?:)?\/\//.test(val)
}

function normalizeAbsolutePath(p) {
  const s = '/' + String(p || '')
  return s.replace(/\/+/g, '/')
}

function normalizeRelativePath(p) {
  return String(p || '').replace(/^\/+/, '')
}

function addTopLevelIfAbsent(r) {
  if (!router.hasRoute(r.name)) {
    router.addRoute(r)
  }
}

function addRouteByChildren(route, segments = [], parentName = null) {
  if (isExternalUrl(route?.path) || isExternalUrl(route?.name) || isExternalUrl(route?.component)) {
    return
  }

  if (route?.name === 'layout') {
    route.children?.forEach((child) => addRouteByChildren(child, [], null))
    return
  }

  if (route?.meta?.defaultMenu === true && parentName === null) {
    const fullPath = [...segments, route.path].filter(Boolean).join('/')
    const children = route.children ? [...route.children] : []
    const newRoute = { ...route, path: fullPath }
    delete newRoute.children
    delete newRoute.parent
    newRoute.path = normalizeAbsolutePath(newRoute.path)

    if (router.hasRoute(newRoute.name)) return
    addTopLevelIfAbsent(newRoute)

    if (children.length) {
      children.forEach((child) => addRouteByChildren(child, [], newRoute.name))
    }
    return
  }

  if (route?.children && route.children.length) {
    if (!parentName) {
      const firstChild = route.children[0]
      if (firstChild) {
        const fullParentPath = [...segments, route.path].filter(Boolean).join('/')
        const redirectPath = normalizeRelativePath(
          [fullParentPath, firstChild.path].filter(Boolean).join('/')
        )
        const parentRoute = {
          path: normalizeRelativePath(fullParentPath),
          name: route.name,
          meta: route.meta,
          redirect: '/layout/' + redirectPath
        }
        router.addRoute('layout', parentRoute)
      }
    }
    const nextSegments = isExternalUrl(route.path) ? segments : [...segments, route.path]
    route.children.forEach((child) => addRouteByChildren(child, nextSegments, parentName))
    return
  }

  const fullPath = [...segments, route.path].filter(Boolean).join('/')
  const newRoute = { ...route, path: fullPath }
  delete newRoute.children
  delete newRoute.parent
  newRoute.path = normalizeRelativePath(newRoute.path)

  if (parentName) {
    router.addRoute(parentName, newRoute)
  } else {
    router.addRoute('layout', newRoute)
  }
}

export function registerAsyncRoutes(baseRouters = []) {
  const layoutRoute = baseRouters[0]
  if (layoutRoute?.name === 'layout' && !router.hasRoute('layout')) {
    const bareLayout = { ...layoutRoute, children: [] }
    router.addRoute(bareLayout)
  }

  const toRegister = []
  if (layoutRoute?.children?.length) {
    toRegister.push(...layoutRoute.children)
  }
  if (baseRouters.length > 1) {
    baseRouters.slice(1).forEach((r) => {
      if (r?.name !== 'layout') toRegister.push(r)
    })
  }
  toRegister.forEach((r) => addRouteByChildren(r, [], null))
}

export function resolveLoginRoute(routeMap = {}, defaultRouter) {
  if (defaultRouter && router.hasRoute(defaultRouter)) {
    return { name: defaultRouter }
  }

  const routes = Object.values(routeMap)
  const preferred = routes.find(
    (item) =>
      item?.name &&
      item.name !== 'Reload' &&
      !item.hidden &&
      router.hasRoute(item.name)
  )
  if (preferred) {
    console.warn(
      `[Login] defaultRouter "${defaultRouter || ''}" 不可用，已回退到 "${preferred.name}"`
    )
    return { name: preferred.name }
  }

  const fallback = routes.find(
    (item) => item?.name && item.name !== 'Reload' && router.hasRoute(item.name)
  )
  if (fallback) {
    console.warn(`[Login] 已回退到首个可用路由 "${fallback.name}"`)
    return { name: fallback.name }
  }

  return null
}
