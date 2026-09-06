package service

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"megin/internal/config"
	systemDto "megin/internal/system/dto"
	"megin/pkg/context/api"
	"megin/pkg/errs"
)

const (
	databaseQueryTimeout = 10 * time.Second
	databaseQueryMaxRows = 1000
)

var forbiddenQueryTokens = map[string]struct{}{
	"ALTER": {}, "ANALYZE": {}, "BENCHMARK": {}, "CALL": {}, "CREATE": {},
	"DELETE": {}, "DO": {}, "DROP": {}, "DUMPFILE": {}, "GET_LOCK": {},
	"GRANT": {}, "HANDLER": {}, "INSERT": {}, "INTO": {}, "LOAD": {},
	"LOCK": {}, "OPTIMIZE": {}, "OUTFILE": {}, "RELEASE_LOCK": {}, "RENAME": {},
	"REPAIR": {}, "REPLACE": {}, "REVOKE": {}, "SET": {}, "SLEEP": {},
	"TRUNCATE": {}, "UNLOCK": {}, "UPDATE": {}, "USE": {},
}

type DatabaseQuery struct {
	ctx *api.Context
}

func NewDatabaseQuery(ctx *api.Context) *DatabaseQuery {
	return &DatabaseQuery{ctx: ctx}
}

func (s *DatabaseQuery) Tables() ([]systemDto.DatabaseTable, error) {
	db, err := s.sqlDB()
	if err != nil {
		return nil, err
	}
	queryCtx, cancel := context.WithTimeout(s.ctx.RequestContext(), databaseQueryTimeout)
	defer cancel()

	rows, err := db.QueryContext(queryCtx, `
		SELECT tables.TABLE_NAME, COALESCE(tables.TABLE_COMMENT, ''), COALESCE(tables.TABLE_ROWS, 0),
		       COALESCE((
		         SELECT keys_table.COLUMN_NAME
		         FROM information_schema.KEY_COLUMN_USAGE AS keys_table
		         WHERE keys_table.TABLE_SCHEMA = tables.TABLE_SCHEMA
		           AND keys_table.TABLE_NAME = tables.TABLE_NAME
		           AND keys_table.CONSTRAINT_NAME = 'PRIMARY'
		         ORDER BY keys_table.ORDINAL_POSITION
		         LIMIT 1
		       ), '') AS order_column
		FROM information_schema.TABLES AS tables
		WHERE tables.TABLE_SCHEMA = DATABASE()
		ORDER BY tables.TABLE_NAME`)
	if err != nil {
		return nil, errs.NewBusinessError(4000, "读取数据表失败："+err.Error())
	}
	defer rows.Close()

	tables := make([]systemDto.DatabaseTable, 0)
	for rows.Next() {
		var table systemDto.DatabaseTable
		if err := rows.Scan(&table.Name, &table.Comment, &table.Rows, &table.OrderColumn); err != nil {
			return nil, errs.NewBusinessError(4000, "读取数据表失败："+err.Error())
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, errs.NewBusinessError(4000, "读取数据表失败："+err.Error())
	}
	return tables, nil
}

func (s *DatabaseQuery) TableStructure(tableName string) (systemDto.DatabaseTableStructure, error) {
	tableName = strings.TrimSpace(tableName)
	if tableName == "" {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(4000, "请选择数据表")
	}
	db, err := s.sqlDB()
	if err != nil {
		return systemDto.DatabaseTableStructure{}, err
	}
	queryCtx, cancel := context.WithTimeout(s.ctx.RequestContext(), databaseQueryTimeout)
	defer cancel()

	var exists int
	if err := db.QueryRowContext(queryCtx, `
		SELECT COUNT(*) FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?`, tableName).Scan(&exists); err != nil {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(4000, "检查数据表失败："+err.Error())
	}
	if exists == 0 {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(404, "数据表不存在")
	}

	quotedTable := "`" + strings.ReplaceAll(tableName, "`", "``") + "`"
	rows, err := db.QueryContext(queryCtx, "SHOW CREATE TABLE "+quotedTable)
	if err != nil {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(4000, "读取表结构失败："+err.Error())
	}
	defer rows.Close()

	columnNames, err := rows.Columns()
	if err != nil {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(4000, "读取表结构失败："+err.Error())
	}
	if !rows.Next() {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(404, "数据表不存在")
	}
	values := make([]any, len(columnNames))
	destinations := make([]any, len(columnNames))
	for i := range values {
		destinations[i] = &values[i]
	}
	if err := rows.Scan(destinations...); err != nil {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(4000, "读取表结构失败："+err.Error())
	}

	createSQL := ""
	for i, columnName := range columnNames {
		if columnName == "Create Table" || columnName == "Create View" {
			createSQL = databaseValueString(values[i])
			break
		}
	}
	if createSQL == "" && len(values) > 1 {
		createSQL = databaseValueString(values[1])
	}
	if createSQL == "" {
		return systemDto.DatabaseTableStructure{}, errs.NewBusinessError(4000, "数据库未返回建表语句")
	}
	return systemDto.DatabaseTableStructure{TableName: tableName, CreateSQL: createSQL}, nil
}

func databaseValueString(value any) string {
	if bytes, ok := value.([]byte); ok {
		return string(bytes)
	}
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func (s *DatabaseQuery) sqlDB() (*sql.DB, error) {
	gormDB := config.GetMysqlDB()
	if gormDB == nil {
		return nil, errs.NewBusinessError(500, "数据库连接未初始化")
	}
	db, err := gormDB.DB()
	if err != nil {
		return nil, errs.NewBusinessError(500, "获取数据库连接失败")
	}
	return db, nil
}

func (s *DatabaseQuery) Execute(query string) (systemDto.DatabaseQueryResult, error) {
	query = strings.TrimSpace(query)
	if err := validateReadOnlyQuery(query); err != nil {
		return systemDto.DatabaseQueryResult{}, errs.NewBusinessError(4000, err.Error())
	}

	db, err := s.sqlDB()
	if err != nil {
		return systemDto.DatabaseQueryResult{}, err
	}

	queryCtx, cancel := context.WithTimeout(s.ctx.RequestContext(), databaseQueryTimeout)
	defer cancel()

	tx, err := db.BeginTx(queryCtx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return systemDto.DatabaseQueryResult{}, errs.NewBusinessError(500, "开启只读事务失败")
	}
	defer func() { _ = tx.Rollback() }()

	startedAt := time.Now()
	rows, err := tx.QueryContext(queryCtx, query)
	if err != nil {
		if queryCtx.Err() != nil {
			return systemDto.DatabaseQueryResult{}, errs.NewBusinessError(4000, "查询超时，最长执行时间为 10 秒")
		}
		return systemDto.DatabaseQueryResult{}, errs.NewBusinessError(4000, "查询执行失败："+err.Error())
	}
	defer rows.Close()

	result, err := collectDatabaseQueryRows(rows)
	if err != nil {
		return systemDto.DatabaseQueryResult{}, errs.NewBusinessError(4000, "读取查询结果失败："+err.Error())
	}
	result.DurationMS = time.Since(startedAt).Milliseconds()
	return result, nil
}

func collectDatabaseQueryRows(rows *sql.Rows) (systemDto.DatabaseQueryResult, error) {
	names, err := rows.Columns()
	if err != nil {
		return systemDto.DatabaseQueryResult{}, err
	}
	types, err := rows.ColumnTypes()
	if err != nil {
		return systemDto.DatabaseQueryResult{}, err
	}

	result := systemDto.DatabaseQueryResult{
		Columns: make([]systemDto.DatabaseQueryColumn, len(names)),
		Rows:    make([][]any, 0),
	}
	for i, name := range names {
		result.Columns[i] = systemDto.DatabaseQueryColumn{Name: name, DatabaseType: types[i].DatabaseTypeName()}
	}

	for rows.Next() {
		if len(result.Rows) >= databaseQueryMaxRows {
			result.Truncated = true
			break
		}
		values := make([]any, len(names))
		destinations := make([]any, len(names))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return systemDto.DatabaseQueryResult{}, err
		}
		for i, value := range values {
			switch typed := value.(type) {
			case []byte:
				values[i] = string(typed)
			case time.Time:
				values[i] = typed.Format(time.RFC3339Nano)
			}
		}
		result.Rows = append(result.Rows, values)
	}
	if err := rows.Err(); err != nil {
		return systemDto.DatabaseQueryResult{}, err
	}
	result.RowCount = len(result.Rows)
	return result, nil
}

type queryToken struct {
	value string
	depth int
}

func validateReadOnlyQuery(query string) error {
	if query == "" {
		return fmt.Errorf("请输入 SQL")
	}
	if strings.Contains(query, "/*!") {
		return fmt.Errorf("不允许使用 MySQL 可执行注释")
	}
	if strings.Contains(query, ":=") {
		return fmt.Errorf("不允许给变量赋值")
	}

	tokens, err := scanQueryTokens(query)
	if err != nil {
		return err
	}
	if len(tokens) == 0 || tokens[0].value != "SELECT" {
		return fmt.Errorf("SQL 必须以 SELECT 开始")
	}

	hasTopLevelLimit := false
	semicolonCount := 0
	for i, token := range tokens {
		if token.value == ";" {
			semicolonCount++
			if i != len(tokens)-1 || semicolonCount > 1 {
				return fmt.Errorf("每次只能执行一条 SELECT 语句")
			}
			continue
		}
		if _, forbidden := forbiddenQueryTokens[token.value]; forbidden {
			return fmt.Errorf("SQL 包含不允许的关键字：%s", token.value)
		}
		if token.value == "LIMIT" && token.depth == 0 {
			hasTopLevelLimit = true
		}
	}
	if !hasTopLevelLimit {
		return fmt.Errorf("SQL 必须包含最外层 LIMIT")
	}
	return nil
}

func scanQueryTokens(query string) ([]queryToken, error) {
	tokens := make([]queryToken, 0, 16)
	depth := 0
	for i := 0; i < len(query); {
		switch query[i] {
		case ' ', '\t', '\r', '\n':
			i++
		case '#':
			i = skipLine(query, i+1)
		case '-':
			if i+1 < len(query) && query[i+1] == '-' {
				i = skipLine(query, i+2)
			} else {
				i++
			}
		case '/':
			if i+1 < len(query) && query[i+1] == '*' {
				end := strings.Index(query[i+2:], "*/")
				if end < 0 {
					return nil, fmt.Errorf("SQL 注释未闭合")
				}
				i += end + 4
			} else {
				i++
			}
		case '\'', '"', '`':
			var err error
			i, err = skipQuoted(query, i, query[i])
			if err != nil {
				return nil, err
			}
		case '(':
			depth++
			i++
		case ')':
			if depth > 0 {
				depth--
			}
			i++
		case ';':
			tokens = append(tokens, queryToken{value: ";", depth: depth})
			i++
		default:
			if isQueryWordByte(query[i]) {
				start := i
				for i < len(query) && isQueryWordByte(query[i]) {
					i++
				}
				tokens = append(tokens, queryToken{value: strings.ToUpper(query[start:i]), depth: depth})
			} else {
				i++
			}
		}
	}
	return tokens, nil
}

func skipLine(query string, start int) int {
	if end := strings.IndexByte(query[start:], '\n'); end >= 0 {
		return start + end + 1
	}
	return len(query)
}

func skipQuoted(query string, start int, quote byte) (int, error) {
	for i := start + 1; i < len(query); i++ {
		if query[i] == '\\' {
			i++
			continue
		}
		if query[i] != quote {
			continue
		}
		if i+1 < len(query) && query[i+1] == quote {
			i++
			continue
		}
		return i + 1, nil
	}
	return 0, fmt.Errorf("SQL 引号未闭合")
}

func isQueryWordByte(value byte) bool {
	return value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' || value == '_' || value == '$'
}
