package validate

import (
	"fmt"
	"megin/pkg/context/api"
	"reflect"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
)

func BindWithPanic(ctx *api.Context, obj interface{}) {
	bindGin(ctx, obj, true)
}

func Bind(ctx *api.Context, obj interface{}) error {
	return bindGin(ctx, obj, false)
}

func bindGin(ctx *api.Context, obj interface{}, withPanic bool) error {
	//绑定参数
	err := ctx.GinCtx.ShouldBind(obj)

	if err != nil {
		//解析错误信息
		value := reflect.TypeOf(obj)
		if errs, ok := err.(validator.ValidationErrors); ok {
			for _, e := range errs {
				if f, exist := value.Elem().FieldByName(e.Field()); exist {
					if tips, ok := f.Tag.Lookup("tips"); ok {
						err = fmt.Errorf("%s", tips)
						break
					} else {
						// 根据验证规则生成友好的错误信息
						switch e.Tag() {
						case "required":
							err = fmt.Errorf("%s字段不能为空", e.Field())
						case "min":
							err = fmt.Errorf("%s字段必须大于等于%s", e.Field(), e.Param())
						case "max":
							err = fmt.Errorf("%s字段必须小于等于%s", e.Field(), e.Param())
						case "url":
							err = fmt.Errorf("%s字段必须是有效的URL", e.Field())
						default:
							err = fmt.Errorf("%s字段验证失败:%s", e.Field(), e.Tag())
						}
						break
					}
				} else {
					err = fmt.Errorf("%s参数绑定错误", e.Field())
					break
				}
			}
		}

		if withPanic {
			//这里的panic将会被中间件捕获
			panic(errors.WithMessage(err, "参数绑定验证失败"))
		}
		return err
	}
	return nil
}

func RegisterExtension() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("datetime", datetime)
		if err != nil {
			panic("ValidatorInstall Error" + err.Error())
		}

		err = v.RegisterValidation("date", date)
		if err != nil {
			panic("ValidatorInstall Error" + err.Error())
		}

		err = v.RegisterValidation("time", time)
		if err != nil {
			panic("ValidatorInstall Error" + err.Error())
		}
	}
}
