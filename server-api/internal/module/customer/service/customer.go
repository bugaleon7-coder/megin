package service

import (
	"megin/internal/base"
	customerConvert "megin/internal/module/customer/convert"
	customerDto "megin/internal/module/customer/dto"
	customerModel "megin/internal/module/customer/model"
	customerRepo "megin/internal/module/customer/repository"
	systemService "megin/internal/system/service"
	"megin/pkg/context/api"
	"megin/pkg/utils"
	"time"
)

type Customer struct {
	base.Service
	Repo *customerRepo.Customer
}

func NewCustomer(ctx *api.Context) *Customer {
	s := &Customer{}
	s.Initialize(ctx)
	s.Repo = customerRepo.NewCustomer(ctx)
	return s
}

func (s *Customer) Create(req *customerDto.CreateCustomerReq, userID, authorityID uint) error {
	now := time.Now()
	customer := customerModel.Customer{
		CustomerName:       req.CustomerName,
		CustomerPhoneData:  req.CustomerPhoneData,
		SysUserID:          userID,
		SysUserAuthorityID: authorityID,
		SystemModel: base.SystemModel{
			CreatedAt: utils.TimePtr(now),
			UpdatedAt: utils.TimePtr(now),
		},
	}
	return s.Repo.Create(&customer)
}

func (s *Customer) Update(req *customerDto.UpdateCustomerReq) error {
	customer, err := s.Repo.GetById(req.ID)
	if err != nil {
		return s.Error(err, "查询客户失败")
	}
	if customer.ID == 0 {
		return s.ErrorMessage("客户不存在")
	}
	customer.CustomerName = req.CustomerName
	customer.CustomerPhoneData = req.CustomerPhoneData
	customer.UpdatedAt = utils.TimePtr(time.Now())
	return s.Repo.Save(&customer)
}

func (s *Customer) Delete(id uint) error {
	customer, err := s.Repo.GetById(id)
	if err != nil {
		return s.Error(err, "查询客户失败")
	}
	if customer.ID == 0 {
		return s.ErrorMessage("客户不存在")
	}
	return s.Repo.DeleteById(id)
}

func (s *Customer) GetByID(id uint) (*customerDto.Customer, error) {
	customer, err := s.Repo.GetByIDWithUser(id)
	if err != nil {
		return nil, s.Error(err, "查询客户失败")
	}
	if customer.ID == 0 {
		return nil, s.ErrorMessage("客户不存在")
	}
	dto := customerConvert.ToCustomerDTO(customer)
	return &dto, nil
}

func (s *Customer) GetList(authorityID uint, req *customerDto.GetCustomerListReq) (*customerDto.PageResult[customerDto.Customer], error) {
	auth, err := systemService.NewSysAuthority(s.Ctx).GetAuthorityInfo(authorityID)
	if err != nil {
		return nil, err
	}

	dataAuthorityIDs := make([]uint, 0, len(auth.DataAuthorityId))
	for i := range auth.DataAuthorityId {
		dataAuthorityIDs = append(dataAuthorityIDs, auth.DataAuthorityId[i].AuthorityId)
	}

	customers, total, err := s.Repo.GetListByAuthorityIDs(dataAuthorityIDs, req.PageNo, req.PageSize)
	if err != nil {
		return nil, s.Error(err, "获取客户列表失败")
	}

	result := &customerDto.PageResult[customerDto.Customer]{
		PageNo:    req.PageNo,
		PageSize:  req.PageSize,
		TotalSize: total,
		List:      make([]customerDto.Customer, len(customers)),
	}
	for i := range customers {
		result.List[i] = customerConvert.ToCustomerDTO(customers[i])
	}
	return result, nil
}
