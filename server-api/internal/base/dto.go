package base

import apiDto "megin/pkg/context/api/dto"

type BaseId struct {
	ID int `form:"id" json:"id"` //ID
}

// Success 表示接口成功且无需返回业务数据的响应。
type Success = apiDto.Success
