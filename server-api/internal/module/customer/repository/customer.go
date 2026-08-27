package repository

import (
	"megin/internal/base"
	customerModel "megin/internal/module/customer/model"
	"megin/pkg/context/api"
)

type Customer struct {
	base.Repository[customerModel.Customer]
}

func NewCustomer(ctx *api.Context) *Customer {
	repo := &Customer{}
	repo.Initialize(ctx)
	return repo
}

func (r *Customer) GetByIDWithUser(id uint) (customerModel.Customer, error) {
	var customer customerModel.Customer
	err := r.DB().Preload("SysUser").Where("id = ?", id).First(&customer).Error
	return customer, err
}

func (r *Customer) GetListByAuthorityIDs(authorityIDs []uint, pageNo, pageSize int) ([]customerModel.Customer, int64, error) {
	customers := make([]customerModel.Customer, 0)
	if len(authorityIDs) == 0 {
		return customers, 0, nil
	}

	query := r.DB().Model(&customerModel.Customer{}).Where("sys_user_authority_id IN ?", authorityIDs)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("SysUser").
		Limit(pageSize).
		Offset((pageNo - 1) * pageSize).
		Find(&customers).Error
	if err != nil {
		return nil, 0, err
	}
	return customers, total, nil
}
