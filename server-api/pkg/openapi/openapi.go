package openapi

import (
	"bytes"
	"fmt"
	"megin/pkg/context/router"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/swaggest/jsonschema-go"
	"github.com/swaggest/openapi-go/openapi3"
)

// ginPathToOpenAPI 将 gin 路径格式 /user/:id 转换为 OpenAPI 格式 /user/{id}
var ginParamRe = regexp.MustCompile(`:(\w+)`)

func ginPathToOpenAPI(path string) string {
	return ginParamRe.ReplaceAllString(path, `{$1}`)
}

type GenerateOptions struct {
	Title       string
	Version     string
	Description string
	OutputPath  string
	RouteFilter func(router.RouteInfo) bool
}

// GenerateOpenAPIDoc 根据路由元数据生成 OpenAPI 文档并写入文件。
// scanDirs 为源码目录，用于 AST 解析注释。
func GenerateOpenAPIDoc(routes []router.RouteInfo, scanDirs []string, options GenerateOptions) error {
	// 解析源码注释
	comments, err := ParseDir(scanDirs)
	if err != nil {
		return fmt.Errorf("解析源码注释失败: %w", err)
	}

	reflector := openapi3.NewReflector()
	reflector.Spec = &openapi3.Spec{Openapi: "3.0.3"}
	if options.Title == "" {
		options.Title = "API 文档"
	}
	if options.Version == "" {
		options.Version = "1.0.0"
	}
	reflector.Spec.Info.
		WithTitle(options.Title).
		WithVersion(options.Version).
		WithDescription(options.Description)

	// schemaToStruct：在 InterceptSchema(Processed=false) 时记录 *Schema 指针 → 结构体名。
	// 同一个 *Schema 指针在 InterceptProp 中作为 ParentSchema 出现，
	// 从而能精确定位字段属于哪个结构体，避免不同结构体同名字段串在一起。
	schemaToStruct := make(map[*jsonschema.Schema]string)

	defNameSanitizerRe := regexp.MustCompile(`[^a-zA-Z0-9.\-_]+`)
	desiredDefNames := map[string]string{}

	reflector.DefaultOptions = append(reflector.DefaultOptions,
		// 非泛型类型直接用类型名（去掉包前缀）；泛型类型生成 Result[dto.UserResp] 格式
		jsonschema.InterceptDefName(func(t reflect.Type, defaultDefName string) string {
			name := t.Name()
			if !strings.Contains(name, "[") {
				pkgPath := t.PkgPath()
				pkgName := pkgPath
				if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
					pkgName = pkgPath[idx+1:]
				}
				if pkgName != "" {
					return pkgName + "." + name
				}
				return name
			}
			// 去掉 Go 泛型运行时在类型参数后追加的 ·N 内部编号
			for {
				idx := strings.IndexRune(name, '·')
				if idx < 0 {
					break
				}
				end := idx + len("·")
				for end < len(name) && name[end] >= '0' && name[end] <= '9' {
					end++
				}
				name = name[:idx] + name[end:]
			}
			// "Result[example.com/m/dto.UserResp]" → "dto.Result[dto.UserResp]"
			bracketIdx := strings.Index(name, "[")
			baseName := name[:bracketIdx]
			typeParam := strings.TrimRight(name[bracketIdx+1:], "]")
			if idx := strings.LastIndex(typeParam, "/"); idx >= 0 {
				typeParam = typeParam[idx+1:]
			}
			pkgPath := t.PkgPath()
			pkgName := pkgPath
			if idx := strings.LastIndex(pkgPath, "/"); idx >= 0 {
				pkgName = pkgPath[idx+1:]
			}
			desired := pkgName + "." + baseName + "[" + typeParam + "]"
			sanitized := defNameSanitizerRe.ReplaceAllString(desired, "")
			desiredDefNames[sanitized] = desired
			return desired
		}),
		// schema 开始处理前，把 *Schema 指针与结构体名绑定（泛型取 base 名）
		jsonschema.InterceptSchema(func(params jsonschema.InterceptSchemaParams) (stop bool, err error) {
			if params.Processed {
				return false, nil
			}
			t := params.Value.Type()
			for t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
			if t.Kind() != reflect.Struct {
				return false, nil
			}
			structName := t.Name()
			// 泛型实例名形如 "Result[main.UserResp]"，只取 base 名与 AST 注释 map 对应
			if i := strings.Index(structName, "["); i >= 0 {
				structName = structName[:i]
			}
			if structName != "" {
				schemaToStruct[params.Schema] = structName
			}
			return false, nil
		}),
		// 属性处理完成后，通过 ParentSchema 精确查找所属结构体的字段注释
		jsonschema.InterceptProp(func(params jsonschema.InterceptPropParams) error {
			if !params.Processed || params.Field.Anonymous {
				return nil
			}
			structName, ok := schemaToStruct[params.ParentSchema]
			if !ok {
				return nil
			}
			if desc := comments.FieldComment(structName, params.Field.Name); desc != "" {
				params.PropertySchema.WithDescription(desc)
			}
			return nil
		}),
	)

	for _, item := range routes {
		if options.RouteFilter != nil && !options.RouteFilter(item) {
			continue
		}

		openAPIPath := ginPathToOpenAPI(item.Path)
		method := strings.ToLower(item.Method)

		op, err := reflector.NewOperationContext(method, openAPIPath)
		if err != nil {
			return fmt.Errorf("创建 operation 失败 [%s %s]: %w", item.Method, item.Path, err)
		}

		// 注入函数注释作为 summary / description
		if item.HandlerName != "" {
			commentKey := item.HandlerName
			if item.ReceiverType != "" {
				commentKey = item.ReceiverType + "." + item.HandlerName
			}
			if s, ok := comments.FuncSummary[commentKey]; ok && s != "" {
				op.SetSummary(s)
			} else if s, ok := comments.FuncSummary[item.HandlerName]; ok && s != "" {
				op.SetSummary(s)
			} else if s, ok := comments.FuncComments[commentKey]; ok && s != "" {
				op.SetSummary(s)
			} else if s, ok := comments.FuncComments[item.HandlerName]; ok && s != "" {
				op.SetSummary(s)
			}
			if d, ok := comments.FuncDescription[commentKey]; ok && d != "" {
				op.SetDescription(d)
			} else if d, ok := comments.FuncDescription[item.HandlerName]; ok && d != "" {
				op.SetDescription(d)
			}
		}

		// 注入 @Tag 作为 OpenAPI 分组
		if item.ReceiverType != "" {
			if tag, ok := comments.StructTags[item.ReceiverType]; ok && tag != "" {
				op.SetTags(tag)
			}
		}

		// 构造请求结构体指针（反射）
		if reqType := reflect.TypeOf(item.ReqType); reqType != nil {
			op.AddReqStructure(reflect.New(reqType).Interface())
		}

		// 构造响应结构体指针（反射）
		if respType := reflect.TypeOf(item.RespType); respType != nil {
			op.AddRespStructure(reflect.New(respType).Interface())
		}

		if err := reflector.AddOperation(op); err != nil {
			return fmt.Errorf("添加 operation 失败 [%s %s]: %w", item.Method, item.Path, err)
		}
	}

	jsonBytes, err := reflector.Spec.MarshalJSON()
	if err != nil {
		return fmt.Errorf("序列化 JSON 失败: %w", err)
	}
	// openapi-go 内部 sanitizeDefName 会去掉 [ ]，在此还原为期望的带括号格式
	for sanitized, desired := range desiredDefNames {
		jsonBytes = bytes.ReplaceAll(jsonBytes, []byte(sanitized), []byte(desired))
	}
	if err := os.MkdirAll(filepath.Dir(options.OutputPath), 0o755); err != nil {
		return fmt.Errorf("创建 OpenAPI 输出目录失败: %w", err)
	}
	return os.WriteFile(options.OutputPath, jsonBytes, 0o644)
}
