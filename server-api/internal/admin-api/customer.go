package handler

import (
	customerDto "megin/internal/module/customer/dto"
	customerService "megin/internal/module/customer/service"
	"megin/pkg/context/api"
)

type Customer struct{}

func (h *Customer) Create(ctx *api.Context, req *customerDto.CreateCustomerReq) (*api.Result[any], error) {
	err := customerService.NewCustomer(ctx).Create(req, uint(ctx.AdminInfo.UserID), uint(ctx.AdminInfo.RoleId))
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func (h *Customer) Update(ctx *api.Context, req *customerDto.UpdateCustomerReq) (*api.Result[any], error) {
	err := customerService.NewCustomer(ctx).Update(req)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func (h *Customer) Delete(ctx *api.Context, req *customerDto.DeleteCustomerReq) (*api.Result[any], error) {
	err := customerService.NewCustomer(ctx).Delete(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultSuccess()
}

func (h *Customer) Detail(ctx *api.Context, req *customerDto.GetCustomerReq) (*api.Result[customerDto.CustomerResponse], error) {
	customer, err := customerService.NewCustomer(ctx).GetByID(req.ID)
	if err != nil {
		return nil, err
	}
	return api.ResultData(customerDto.CustomerResponse{Customer: *customer})
}

func (h *Customer) List(ctx *api.Context, req *customerDto.GetCustomerListReq) (*api.Result[customerDto.PageResult[customerDto.Customer]], error) {
	result, err := customerService.NewCustomer(ctx).GetList(uint(ctx.AdminInfo.RoleId), req)
	if err != nil {
		return nil, err
	}
	return api.ResultData(*result)
}
