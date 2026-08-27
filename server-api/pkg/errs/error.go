package errs

import "fmt"

const DS = "###"

// 业务错误,提示类错误,非程序异常
type BusinessError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewBusinessError(code int, message string) *BusinessError {
	return &BusinessError{code, message}
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("%v%s%v", e.Code, DS, e.Message)
}

func IsBusinessError(err error) (*BusinessError, bool) {
	if err != nil {
		switch err.(type) {
		case *BusinessError:
			return err.(*BusinessError), true
		default:
			return nil, false
		}
	}
	return nil, false
}

// 普通错误,程序异常
type NormalError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     error
}

func NewNormalError(code int, message string, err error) *NormalError {
	return &NormalError{code, message, err}
}

func (e *NormalError) Error() string {
	return fmt.Sprintf("%v%s%v", e.Code, DS, e.Message)
}

func IsNormalError(err error) (*NormalError, bool) {
	if err != nil {
		switch err.(type) {
		case *NormalError:
			return err.(*NormalError), true
		default:
			return nil, false
		}
	}
	return nil, false
}
