package process_test

import (
	"fmt"
	"strings"
	"testing"

	"sqlSyntaxAudit/config"
	"sqlSyntaxAudit/controllers/inspect"
	"sqlSyntaxAudit/controllers/parser"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

// Exercise SQL parsing, column traversal, configuration wiring and validation
// together without the database required by the inspect package's test setup.
func TestColumnSupportCharset(t *testing.T) {
	operations := []struct {
		name string
		sql  string
	}{
		{"create", "CREATE TABLE arch (address %s) DEFAULT CHARSET=utf8mb4"},
		{"add", "ALTER TABLE arch ADD COLUMN address %s"},
		{"modify", "ALTER TABLE arch MODIFY COLUMN address %s"},
		{"change", "ALTER TABLE arch CHANGE COLUMN old_address address %s"},
	}
	tests := []struct {
		name       string
		definition string
		allowed    []string
		disabled   bool
		wantError  string
	}{
		{
			name:       "reject utf8 override",
			definition: "VARCHAR(50) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT 'xx'",
			allowed:    []string{"utf8mb4"},
			wantError:  "列`address`指定的字符集`utf8`不符合要求，可选字符集为[utf8mb4][表`arch`]",
		},
		{
			name:       "allow general collation",
			definition: "VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT 'xx'",
			allowed:    []string{"utf8mb4"},
		},
		{
			name:       "allow binary collation",
			definition: "VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin DEFAULT 'xx'",
			allowed:    []string{"utf8mb4"},
		},
		{
			name:       "allow multiple charsets independently of table whitelist",
			definition: "VARCHAR(50) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT 'xx'",
			allowed:    []string{"utf8", "utf8mb4"},
		},
		{
			name:       "omitted whitelist preserves behavior",
			definition: "VARCHAR(50) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT 'xx'",
		},
		{
			name:       "empty whitelist preserves behavior",
			definition: "VARCHAR(50) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT 'xx'",
			allowed:    []string{},
		},
		{
			name:       "disabled check ignores whitelist",
			definition: "VARCHAR(50) CHARACTER SET utf8 COLLATE utf8_general_ci DEFAULT 'xx'",
			allowed:    []string{"utf8mb4"},
			disabled:   true,
		},
		{
			name:       "case insensitive whitelist",
			definition: "VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT 'xx'",
			allowed:    []string{"UTF8MB4"},
		},
		{
			name:       "implicit charset preserves inheritance",
			definition: "VARCHAR(50) DEFAULT 'xx'",
			allowed:    []string{"utf8mb4"},
		},
		{
			name:       "missing collation still rejected",
			definition: "VARCHAR(50) CHARACTER SET utf8mb4 DEFAULT 'xx'",
			allowed:    []string{"utf8mb4"},
			wantError:  "必须同时指定字符集和排序规则",
		},
		{
			name:       "noncharacter column unaffected",
			definition: "INT DEFAULT 0",
			allowed:    []string{"utf8mb4"},
		},
	}
	for _, operation := range operations {
		for _, tt := range tests {
			t.Run(operation.name+"/"+tt.name, func(t *testing.T) {
				audit, warns, err := parser.NewParse(fmt.Sprintf(operation.sql, tt.definition), "", "")
				if err != nil || len(warns) != 0 {
					t.Fatalf("parse: error=%v, warnings=%v", err, warns)
				}
				rule := &inspect.Rule{AuditConfig: &config.AuditConfiguration{
					CHECK_TABLE_CHARSET:    true,
					TABLE_SUPPORT_CHARSET:  []map[string]string{{"charset": "utf8mb4", "recommend": "utf8mb4_general_ci"}},
					CHECK_COLUMN_CHARSET:   !tt.disabled,
					COLUMN_SUPPORT_CHARSET: tt.allowed,
				}}
				stmt := audit.TiStmt[0]
				switch stmt.(type) {
				case *ast.CreateTableStmt:
					rule.RuleCreateTableColsCharset(&stmt)
				case *ast.AlterTableStmt:
					rule.RuleAlterTableColCharset(&stmt)
				default:
					t.Fatalf("unexpected statement type %T", stmt)
				}
				if tt.wantError == "" {
					if len(rule.Summary) != 0 {
						t.Fatalf("unexpected audit errors: %v", rule.Summary)
					}
				} else if len(rule.Summary) != 1 || !strings.Contains(rule.Summary[0], tt.wantError) {
					t.Fatalf("audit errors=%v, want one containing %q", rule.Summary, tt.wantError)
				}
			})
		}
	}
}
