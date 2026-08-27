package api

import (
	"fmt"
	err "megin/pkg/errs"
	"strings"
)

type Result[T any] struct {
	Code    int    `json:"code"`     //返回编码200为成功,其它编号为异常
	Message string `json:"message"`  //返回错误信息
	Data    T      `json:"data"`     //返回的数值
	Trace   []string `json:"trace,omitempty"` //错误堆栈明细,按行拆分
	TraceId string `json:"trace_id"` //请求链路ID,可用来方便查询log
	Success bool   `json:"success"`  //是否成功
}

func ResultSuccess() (*Result[any], error) {
	return &Result[any]{
		Code:    STATUS_SUCCESS,
		Message: "成功",
		Success: true,
	}, nil
}

func ResultData[T any](dataArgs ...T) (*Result[T], error) {
	var data T
	if len(dataArgs) > 0 {
		data = dataArgs[0]
	}
	return &Result[T]{
		Code:    STATUS_SUCCESS,
		Message: "成功",
		Success: true,
		Data:    data,
	}, nil
}

func Failed[T any](error error, msgarg ...string) *Result[T] {
	tips := ""
	if len(msgarg) > 0 {
		tips = msgarg[0]
	}

	if bizErr, ok := err.IsBusinessError(error); ok {
		if len(tips) > 0 {
			bizErr.Message = tips
		}

		return &Result[T]{
			Code:    bizErr.Code,
			Message: bizErr.Message,
			Trace:   normalizeTrace(fmt.Sprintf("%+v", error)),
			Success: false,
		}
	}

	if normalErr, ok := err.IsNormalError(error); ok {
		if len(tips) > 0 {
			normalErr.Message = tips
		}

		return &Result[T]{
			Code:    normalErr.Code,
			Message: normalErr.Message,
			Trace:   normalizeTrace(fmt.Sprintf("%+v", normalErr.Err)),
			Success: false,
		}
	}

	if len(tips) == 0 {
		tips = error.Error()
	}

	return &Result[T]{
		Code:    STATUS_SERVER_ERROR,
		Message: tips,
		Trace:   normalizeTrace(fmt.Sprintf("%+v", error)),
		Success: false,
	}
}

// 统一清洗错误堆栈行，避免接口直接返回带空白缩进的内容。
func normalizeTrace(trace string) []string {
	lines := strings.Split(strings.TrimRight(trace, "\n"), "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		result = append(result, line)
	}
	return result
}
