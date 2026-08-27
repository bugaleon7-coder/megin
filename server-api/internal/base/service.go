package base

import (
	"megin/pkg/context/api"
	"megin/pkg/errs"
	"megin/pkg/logger"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Service struct {
	Ctx *api.Context
	log *logger.Logger
}

func (s *Service) Initialize(ctx *api.Context) *Service {
	s.Ctx = ctx
	s.log = ctx.Log
	return s
}

func (s *Service) ErrorMessage(message string, code ...int) error {
	if len(code) > 0 {
		return errs.NewBusinessError(code[0], message)
	}
	return errs.NewBusinessError(500, message)
}

func (s *Service) ErrorCodeMessage(code int, message string) error {
	return errs.NewBusinessError(code, message)
}

func (s *Service) Error(err error, message ...string) error {
	if err == nil {
		return err
	}

	s.log.Info("error", zap.Any("message", errors.WithStack(err)))
	if len(message) > 0 {
		return errs.NewNormalError(500, message[0], errors.WithStack(err))
	}
	return errs.NewNormalError(500, "服务器错误:"+err.Error(), errors.WithStack(err))
}

func (s *Service) Errorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return err
	}
	return errors.Wrapf(err, format, args...)
}
