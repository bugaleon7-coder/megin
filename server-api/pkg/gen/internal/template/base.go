package template

const NotEditMark = `
// 该文件由tools/gorm_gen.go生成
`

const Header = NotEditMark + `
package {{.Package}}

import(	
	{{range .ImportPkgPaths}}{{.}}` + "\n" + `{{end}}
)
`
