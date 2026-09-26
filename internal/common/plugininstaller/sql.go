package plugininstaller

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var tableStatementPattern = regexp.MustCompile(`(?is)^\s*(?:create\s+table(?:\s+if\s+not\s+exists)?|alter\s+table|insert\s+into)\s+\x60?([a-zA-Z0-9_]+)\x60?`)
var createIndexPattern = regexp.MustCompile(`(?is)^\s*create\s+(?:unique\s+)?index\s+\x60?[a-zA-Z0-9_]+\x60?\s+on\s+\x60?([a-zA-Z0-9_]+)\x60?`)
var pluginTableReferencePattern = regexp.MustCompile(`(?i)\bplg_[a-z0-9_]+`)
var insertColumnsPattern = regexp.MustCompile(`(?is)^\s*insert\s+into\s+\x60?[a-zA-Z0-9_]+\x60?\s*\(([^)]*)\)`)
var destructiveAlterPattern = regexp.MustCompile(`(?i)\b(drop|change|modify|rename)\b`)

// ValidateMigrationSQL 限制插件迁移只能操作自己的 plg_{plugin_id}_ 表。
func ValidateMigrationSQL(pluginID, content string) error {
	statements, err := SplitSQLStatements(content)
	if err != nil {
		return err
	}
	if len(statements) == 0 {
		return errors.New("迁移文件没有可执行语句")
	}
	prefix := "plg_" + pluginID + "_"
	for _, statement := range statements {
		lower := strings.ToLower(strings.TrimSpace(statement))
		if strings.Contains(lower, "sys_") || strings.Contains(lower, "outfile") || strings.Contains(lower, "dumpfile") {
			return errors.New("迁移不允许引用核心表或文件导入导出语句")
		}
		if strings.Contains(lower, "on duplicate key") {
			return errors.New("迁移不允许使用ON DUPLICATE KEY覆盖现有数据")
		}
		matches := tableStatementPattern.FindStringSubmatch(statement)
		if len(matches) == 0 {
			matches = createIndexPattern.FindStringSubmatch(statement)
		}
		if len(matches) < 2 {
			return errors.New("只允许CREATE TABLE、非破坏性ALTER TABLE、CREATE INDEX和INSERT")
		}
		if !strings.HasPrefix(strings.ToLower(matches[1]), prefix) {
			return fmt.Errorf("表%s不属于插件前缀%s", matches[1], prefix)
		}
		for _, referencedTable := range pluginTableReferencePattern.FindAllString(lower, -1) {
			if !strings.HasPrefix(referencedTable, prefix) {
				return fmt.Errorf("迁移引用了其他插件表%s", referencedTable)
			}
		}
		if strings.HasPrefix(lower, "alter table") {
			if destructiveAlterPattern.MatchString(lower) {
				return errors.New("ALTER TABLE只允许新增字段或索引")
			}
		}
		if strings.HasPrefix(lower, "insert into") {
			columnMatches := insertColumnsPattern.FindStringSubmatch(statement)
			if len(columnMatches) < 2 {
				return errors.New("INSERT必须显式声明字段列表")
			}
			for _, column := range strings.Split(columnMatches[1], ",") {
				if strings.Trim(strings.TrimSpace(strings.ToLower(column)), "`") == "id" {
					return errors.New("INSERT不允许指定自增ID")
				}
			}
		}
	}
	return nil
}

// SplitSQLStatements 按分号切分SQL，并正确处理引号和注释中的分号。
func SplitSQLStatements(content string) ([]string, error) {
	var statements []string
	var builder strings.Builder
	var quote rune
	escaped := false
	lineComment := false
	blockComment := false
	runes := []rune(content)
	for index := 0; index < len(runes); index++ {
		current := runes[index]
		next := rune(0)
		if index+1 < len(runes) {
			next = runes[index+1]
		}
		if lineComment {
			if current == '\n' {
				lineComment = false
				builder.WriteRune(' ')
			}
			continue
		}
		if blockComment {
			if current == '*' && next == '/' {
				blockComment = false
				index++
				builder.WriteRune(' ')
			}
			continue
		}
		if quote == 0 {
			if current == '#' || (current == '-' && next == '-' && (index+2 >= len(runes) || unicode.IsSpace(runes[index+2]))) {
				lineComment = true
				if current == '-' {
					index++
				}
				continue
			}
			if current == '/' && next == '*' {
				blockComment = true
				index++
				continue
			}
			if current == '\'' || current == '"' || current == '`' {
				quote = current
				builder.WriteRune(current)
				continue
			}
			if current == ';' {
				if statement := strings.TrimSpace(builder.String()); statement != "" {
					statements = append(statements, statement)
				}
				builder.Reset()
				continue
			}
			builder.WriteRune(current)
			continue
		}
		builder.WriteRune(current)
		if escaped {
			escaped = false
			continue
		}
		if current == '\\' && quote != '`' {
			escaped = true
			continue
		}
		if current == quote {
			if next == quote {
				builder.WriteRune(next)
				index++
				continue
			}
			quote = 0
		}
	}
	if quote != 0 || blockComment {
		return nil, errors.New("SQL存在未闭合的引号或注释")
	}
	if statement := strings.TrimSpace(builder.String()); statement != "" {
		statements = append(statements, statement)
	}
	return statements, nil
}
