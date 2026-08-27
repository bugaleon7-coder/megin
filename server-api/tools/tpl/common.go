package tpl

type CommonMethod struct {
	ID int64
}

// 计录是否存在
func (m CommonMethod) isNil() bool {
	return m.ID == 0
}

// 获取ID
func (m CommonMethod) GetID() any {
	return m.ID
}
