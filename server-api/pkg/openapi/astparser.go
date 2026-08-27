package openapi

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// Comments holds all extracted comment information from source files.
type Comments struct {
	// FuncComments 映射函数注释，key 支持 "方法名" 或 "接收者类型.方法名"。
	FuncComments map[string]string
	// FuncSummary 映射 @Summary，key 支持 "方法名" 或 "接收者类型.方法名"。
	FuncSummary map[string]string
	// FuncDescription 映射 @Description，key 支持 "方法名" 或 "接收者类型.方法名"。
	FuncDescription map[string]string
	// StructComments maps struct name -> doc comment text
	StructComments map[string]string
	// FieldComments maps struct name -> field name -> comment text
	FieldComments map[string]map[string]string
	// EmbeddedStructs maps struct name -> list of embedded (anonymous) struct type names
	EmbeddedStructs map[string][]string
	// StructTags maps struct name -> @Tag value (OpenAPI tag/group name)
	StructTags map[string]string
}

// ParseDir parses all .go files in the given directory (recursively) and extracts comments.
func ParseDir(dirs []string) (*Comments, error) {
	c := &Comments{
		FuncComments:    make(map[string]string),
		FuncSummary:     make(map[string]string),
		FuncDescription: make(map[string]string),
		StructComments:  make(map[string]string),
		FieldComments:   make(map[string]map[string]string),
		EmbeddedStructs: make(map[string][]string),
		StructTags:      make(map[string]string),
	}

	fset := token.NewFileSet()
	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			// 跳过隐藏目录和 vendor（根目录本身不跳过）
			base := filepath.Base(path)
			if path != dir && (strings.HasPrefix(base, ".") || base == "vendor") {
				return filepath.SkipDir
			}
			pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
			if err != nil {
				return err
			}
			for _, pkg := range pkgs {
				for _, file := range pkg.Files {
					c.extractFile(file)
				}
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

// ParseFiles parses the given source files and extracts comments.
func ParseFiles(files ...string) (*Comments, error) {
	fset := token.NewFileSet()

	c := &Comments{
		FuncComments:    make(map[string]string),
		FuncSummary:     make(map[string]string),
		FuncDescription: make(map[string]string),
		StructComments:  make(map[string]string),
		FieldComments:   make(map[string]map[string]string),
		EmbeddedStructs: make(map[string][]string),
		StructTags:      make(map[string]string),
	}

	for _, path := range files {
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		c.extractFile(f)
	}

	return c, nil
}

// FieldComment returns the comment for the given struct field, searching the struct itself
// and any embedded structs (recursively).
func (c *Comments) FieldComment(structName, fieldName string) string {
	return c.fieldCommentDeep(structName, fieldName, map[string]bool{})
}

func (c *Comments) fieldCommentDeep(structName, fieldName string, visited map[string]bool) string {
	if visited[structName] {
		return ""
	}
	visited[structName] = true

	// Check direct fields of this struct first
	if fieldMap, ok := c.FieldComments[structName]; ok {
		if desc, ok := fieldMap[fieldName]; ok && desc != "" {
			return desc
		}
	}
	// Recurse into embedded structs
	for _, embedded := range c.EmbeddedStructs[structName] {
		if desc := c.fieldCommentDeep(embedded, fieldName, visited); desc != "" {
			return desc
		}
	}
	return ""
}

func (c *Comments) extractFile(f *ast.File) {
	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			c.extractFunc(d)
		case *ast.GenDecl:
			c.extractGenDecl(d)
		}
	}
}

func (c *Comments) extractFunc(fn *ast.FuncDecl) {
	if fn.Doc == nil {
		return
	}
	text := fn.Doc.Text()
	name := fn.Name.Name
	qualifiedName := name
	if receiver := receiverTypeName(fn); receiver != "" {
		qualifiedName = receiver + "." + name
	}
	summary, desc, plain := parseAnnotations(text)
	if summary != "" {
		c.FuncSummary[name] = summary
		c.FuncSummary[qualifiedName] = summary
	}
	if desc != "" {
		c.FuncDescription[name] = desc
		c.FuncDescription[qualifiedName] = desc
	}
	if plain != "" {
		c.FuncComments[name] = plain
		c.FuncComments[qualifiedName] = plain
	}
}

// receiverTypeName 提取方法接收者类型名，供同名方法注释精确匹配使用。
func receiverTypeName(fn *ast.FuncDecl) string {
	if fn.Recv == nil || len(fn.Recv.List) == 0 {
		return ""
	}
	return embeddedTypeName(fn.Recv.List[0].Type)
}

func (c *Comments) extractGenDecl(decl *ast.GenDecl) {
	for _, spec := range decl.Specs {
		ts, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		st, ok := ts.Type.(*ast.StructType)
		if !ok {
			continue
		}

		structName := ts.Name.Name

		// struct-level doc: prefer spec doc, fall back to group doc
		docText := ""
		if ts.Doc != nil {
			docText = ts.Doc.Text()
			c.StructComments[structName] = cleanComment(docText)
		} else if decl.Doc != nil {
			docText = decl.Doc.Text()
			c.StructComments[structName] = cleanComment(docText)
		}
		// 解析 @Tag 注解
		if tag := parseAtTag(docText); tag != "" {
			c.StructTags[structName] = tag
		}

		fieldMap := make(map[string]string)
		var embeddedTypes []string

		for _, field := range st.Fields.List {
			if len(field.Names) == 0 {
				// Anonymous/embedded field
				typeName := embeddedTypeName(field.Type)
				if typeName != "" {
					embeddedTypes = append(embeddedTypes, typeName)
				}
				continue
			}
			comment := fieldComment(field)
			if comment == "" {
				continue
			}
			for _, name := range field.Names {
				fieldMap[name.Name] = comment
			}
		}

		if len(fieldMap) > 0 {
			c.FieldComments[structName] = fieldMap
		}
		if len(embeddedTypes) > 0 {
			c.EmbeddedStructs[structName] = embeddedTypes
		}
	}
}

// embeddedTypeName extracts the type name from an anonymous (embedded) field type expression.
func embeddedTypeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return embeddedTypeName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	}
	return ""
}

// fieldComment returns the comment for a struct field.
// Priority: doc comment (above) > line comment (inline).
func fieldComment(field *ast.Field) string {
	if field.Doc != nil && field.Doc.Text() != "" {
		return cleanComment(field.Doc.Text())
	}
	if field.Comment != nil && field.Comment.Text() != "" {
		return cleanComment(field.Comment.Text())
	}
	return ""
}

// cleanComment trims whitespace and trailing newlines from a comment block.
func cleanComment(s string) string {
	return strings.TrimSpace(s)
}

// parseAtTag extracts the value of "@Tag <name>" from a doc comment string.
func parseAtTag(comment string) string {
	for _, line := range strings.Split(comment, "\n") {
		if value := extractAnnotationValue(line, "@Tag"); value != "" {
			return value
		}
	}
	return ""
}

// parseAnnotations splits a doc comment into @Summary, @Description, and remaining plain text.
func parseAnnotations(comment string) (summary, description, plain string) {
	var plainLines []string
	var descriptions []string

	for _, line := range strings.Split(comment, "\n") {
		if value := extractAnnotationValue(line, "@Summary"); value != "" {
			summary = value
			continue
		}
		if value := extractAnnotationValue(line, "@Description"); value != "" {
			descriptions = append(descriptions, value)
			continue
		}
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			plainLines = append(plainLines, trimmed)
		}
	}
	plain = strings.TrimSpace(strings.Join(plainLines, "\n"))
	description = strings.TrimSpace(strings.Join(descriptions, "<br />"))
	return
}

// extractAnnotationValue returns the text after the first occurrence of marker in line.
// Any prefix before the marker is ignored.
func extractAnnotationValue(line, marker string) string {
	idx := strings.Index(line, marker)
	if idx < 0 {
		return ""
	}
	value := strings.TrimSpace(line[idx+len(marker):])
	if value == "" {
		return ""
	}
	return value
}
