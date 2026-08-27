import service from '@/utils/request'
// @Tags UserInfo
// @Summary 创建userInfo表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.UserInfo true "创建userInfo表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /userInfo/createUserInfo [post]
export const createUserInfo = (data) => {
  return service({
    url: '/userInfo/createUserInfo',
    method: 'post',
    data
  })
}

// @Tags UserInfo
// @Summary 删除userInfo表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.UserInfo true "删除userInfo表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /userInfo/deleteUserInfo [delete]
export const deleteUserInfo = (params) => {
  return service({
    url: '/userInfo/deleteUserInfo',
    method: 'delete',
    params
  })
}

// @Tags UserInfo
// @Summary 批量删除userInfo表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body request.IdsReq true "批量删除userInfo表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /userInfo/deleteUserInfo [delete]
export const deleteUserInfoByIds = (params) => {
  return service({
    url: '/userInfo/deleteUserInfoByIds',
    method: 'delete',
    params
  })
}

// @Tags UserInfo
// @Summary 更新userInfo表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data body model.UserInfo true "更新userInfo表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /userInfo/updateUserInfo [put]
export const updateUserInfo = (data) => {
  return service({
    url: '/userInfo/updateUserInfo',
    method: 'put',
    data
  })
}

// @Tags UserInfo
// @Summary 用id查询userInfo表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query model.UserInfo true "用id查询userInfo表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /userInfo/findUserInfo [get]
export const findUserInfo = (params) => {
  return service({
    url: '/userInfo/findUserInfo',
    method: 'get',
    params
  })
}

// @Tags UserInfo
// @Summary 分页获取userInfo表列表
// @Security ApiKeyAuth
// @Accept application/json
// @Produce application/json
// @Param data query request.PageInfo true "分页获取userInfo表列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /userInfo/getUserInfoList [get]
export const getUserInfoList = (params) => {
  return service({
    url: '/userInfo/getUserInfoList',
    method: 'get',
    params
  })
}

// @Tags UserInfo
// @Summary 不需要鉴权的userInfo表接口
// @Accept application/json
// @Produce application/json
// @Param data query exampleReq.UserInfoSearch true "分页获取userInfo表列表"
// @Success 200 {object} response.Response{data=object,msg=string} "获取成功"
// @Router /userInfo/getUserInfoPublic [get]
export const getUserInfoPublic = () => {
  return service({
    url: '/userInfo/getUserInfoPublic',
    method: 'get',
  })
}
