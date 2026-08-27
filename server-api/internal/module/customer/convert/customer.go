package convert

import (
	customerDto "megin/internal/module/customer/dto"
	customerModel "megin/internal/module/customer/model"
)

func ToCustomerDTO(customer customerModel.Customer) customerDto.Customer {
	return customerDto.Customer{
		ID:                 customer.ID,
		CreatedAt:          customer.CreatedAt,
		UpdatedAt:          customer.UpdatedAt,
		CustomerName:       customer.CustomerName,
		CustomerPhoneData:  customer.CustomerPhoneData,
		SysUserID:          customer.SysUserID,
		SysUserAuthorityID: customer.SysUserAuthorityID,
		SysUser:            customer.SysUser,
	}
}
