package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"megin/pkg/bootstrap"
	"megin/pkg/utils"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"

	"github.com/spf13/cast"
)

var (
	currentToken string
)

// Login 登录并获取token
func Login(username, password string) string {
	if currentToken != "" {
		return currentToken
	}

	resp := PostWithoutToken("/admin-api/user/login", map[string]string{
		"username": username,
		"password": password,
	})
	body := resp.Body.String()

	// 解析响应，获取token
	var loginResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Success bool   `json:"success"`
		Data    struct {
			User      any    `json:"user"`
			Token     string `json:"token"`
			ExpiresAt int64  `json:"expiresAt"`
		} `json:"data"`
	}

	if err := json.Unmarshal([]byte(body), &loginResp); err != nil {
		fmt.Printf("解析登录响应失败: %v\n", err)
		return ""
	}

	if loginResp.Success && loginResp.Data.Token != "" {
		currentToken = loginResp.Data.Token
		fmt.Printf("登录成功，获取到token: %s\n", currentToken)
		return currentToken
	}

	fmt.Printf("登录失败: %s\n", loginResp.Message)
	return currentToken
}

// PostWithoutToken 不带token的POST请求（用于登录）
func PostWithoutToken(path string, params interface{}) *httptest.ResponseRecorder {
	r := bootstrap.SetupTestRouter()
	body, _ := json.Marshal(params)
	fmt.Println("请求Path:", path)
	//fmt.Println("请求参数:", string(body))
	req := httptest.NewRequest("POST", path, bytes.NewReader(body))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Body.String()
	fmt.Println("返回数据:", resp)
	return w
}

func Post(path string, params interface{}) *httptest.ResponseRecorder {
	// 确保已登录并获取了token
	//Login("admin", "admin")

	r := bootstrap.SetupTestRouter()
	body, _ := json.Marshal(params)
	fmt.Println("请求Path:", path)
	//fmt.Println("请求参数:", string(body))
	req := httptest.NewRequest("POST", path, bytes.NewReader(body))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	req.Header.Add("Token", currentToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	resp := w.Body.String()
	fmt.Println("返回数据:", resp)
	return w
}

// r := system.SetupTestRouter()
// req := httptest.NewRequest("GET", path+"?"+querystring, strings.NewReader(""))
// //json格式提交
// req.Header.Add("Content-type", "application/json;charset=utf-8")
// req.Header.Add("platform", "PH")
// w := httptest.NewRecorder()
// r.ServeHTTP(w, req)
// return w
func GetWithToken(path string, token string, params any) *httptest.ResponseRecorder {
	r := bootstrap.SetupTestRouter()
	var querystring string
	var err error
	if params != nil {
		querystring, err = utils.HttpBuildQuery(params, false)
		if err != nil {
			panic("参数不正确:" + err.Error())
		}
	}

	req := httptest.NewRequest("GET", path+"?"+querystring, strings.NewReader(querystring))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	req.Header.Add("Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Get(path string, params any) *httptest.ResponseRecorder {
	// 确保已登录并获取了token
	//Login("admin", "admin")

	r := bootstrap.SetupTestRouter()
	var querystring string
	var err error
	if params != nil {
		querystring, err = utils.HttpBuildQuery(params, false)
		if err != nil {
			panic("参数不正确:" + err.Error())
		}
	}

	req := httptest.NewRequest("GET", path+"?"+querystring, strings.NewReader(querystring))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	req.Header.Add("Token", currentToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Delete(path string) *httptest.ResponseRecorder {
	// 确保已登录并获取了token
	Login("admin", "admin")

	r := bootstrap.SetupTestRouter()
	req := httptest.NewRequest("DELETE", path, strings.NewReader(""))
	//json格式提交
	req.Header.Add("Content-type", "application/json;charset=utf-8")
	req.Header.Add("Token", currentToken)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func Print(data interface{}) {
	t := reflect.TypeOf(data).String()
	var res []byte
	if t == "string" {
		res = []byte(cast.ToString(data))
	} else {
		res, _ = json.Marshal(data)
	}

	var out bytes.Buffer
	_ = json.Indent(&out, res, "", "\t")
	out.WriteTo(os.Stdout)
	fmt.Printf("\n")
}
