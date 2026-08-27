package base

import (
	"time"

	"gorm.io/gorm"
)

type Model interface {
	TableName() string
	GetID() any
}

type ControlBy struct {
	CreateBy int `json:"create_by" gorm:"comment:创建者"`
	UpdateBy int `json:"update_by" gorm:"comment:更新者"`
}

// SetCreateBy 设置创建人id
func (e *ControlBy) SetCreateBy(createBy int) {
	e.CreateBy = createBy
}

// SetUpdateBy 设置修改人id
func (e *ControlBy) SetUpdateBy(updateBy int) {
	e.UpdateBy = updateBy
}

type SystemModel struct {
	ID        uint           `gorm:"primarykey" json:"ID"`
	CreatedAt *time.Time     `gorm:"column:created_at;comment:创建时间"`
	UpdatedAt *time.Time     `gorm:"column:updated_at;comment:最后更新时间"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
