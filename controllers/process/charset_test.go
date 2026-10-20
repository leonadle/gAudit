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

func collationAudit(t *testing.T, sql string, cfg *config.AuditConfiguration) []string {
	t.Helper()
	audit, warns, err := parser.NewParse(sql, "", "")
	if err != nil || len(warns) != 0 {
		t.Fatalf("parse: error=%v, warnings=%v", err, warns)
	}
	rule := &inspect.Rule{AuditConfig: cfg}
	stmt := audit.TiStmt[0]
	switch stmt.(type) {
	case *ast.CreateTableStmt:
		rule.RuleCreateTableOptions(&stmt)
		rule.RuleCreateTableColsCharset(&stmt)
	case *ast.AlterTableStmt:
		rule.RuleAlterTableOptions(&stmt)
		rule.RuleAlterTableColCharset(&stmt)
	default:
		t.Fatalf("unexpected statement type %T", stmt)
	}
	return rule.Summary
}

func TestExplicitCollationWhitelist(t *testing.T) {
	tests := []struct {
		name      string
		sql       string
		wantError string
	}{
		{"general", "DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci", ""},
		{"bin", "DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin", ""},
		{"unicode", "DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci", "表`arch`指定的排序规则`utf8mb4_unicode_ci`不符合要求"},
		{"implicit", "DEFAULT CHARSET=utf8mb4", ""},
		{"utf8", "DEFAULT CHARSET=utf8 COLLATE=utf8_general_ci", "表`arch`指定的字符集`utf8`不符合要求"},
	}
	for _, operation := range []string{"CREATE TABLE arch (id INT) ENGINE=InnoDB ", "ALTER TABLE arch "} {
		for _, tt := range tests {
			t.Run(operation+tt.name, func(t *testing.T) {
				assertCollationAudit(t, operation+tt.sql, collationConfig(), tt.wantError)
			})
		}
	}

	columnTests := []struct {
		name       string
		definition string
		wantError  string
	}{
		{"general", "VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci", ""},
		{"bin", "VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_bin", ""},
		{"unicode", "VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci", "列`address`指定的排序规则`utf8mb4_unicode_ci`不符合要求"},
		{"charset only", "VARCHAR(50) CHARACTER SET utf8mb4", ""},
		{"collate only", "VARCHAR(50) COLLATE utf8mb4_bin", ""},
		{"implicit", "VARCHAR(50)", ""},
		{"disallowed charset without collate", "VARCHAR(50) CHARACTER SET utf8", "列`address`指定的字符集`utf8`不符合要求"},
		{"text unicode", "TEXT COLLATE utf8mb4_unicode_ci", "列`address`指定的排序规则`utf8mb4_unicode_ci`不符合要求"},
		{"text bin", "TEXT COLLATE utf8mb4_bin", ""},
		{"implicit text", "TEXT", ""},
		{"char", "CHAR(50) COLLATE utf8mb4_general_ci", ""},
		{"enum", "ENUM('a', 'b') COLLATE utf8mb4_bin", ""},
		{"set", "SET('a', 'b') COLLATE utf8mb4_bin", ""},
		{"implicit json", "JSON", ""},
		{"implicit blob", "BLOB", ""},
	}
	for _, operation := range []string{
		"CREATE TABLE arch (address %s) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4",
		"ALTER TABLE arch ADD COLUMN address %s",
		"ALTER TABLE arch MODIFY COLUMN address %s",
		"ALTER TABLE arch CHANGE COLUMN old_address address %s",
	} {
		for _, tt := range columnTests {
			t.Run(operation+tt.name, func(t *testing.T) {
				assertCollationAudit(t, fmt.Sprintf(operation, tt.definition), collationConfig(), tt.wantError)
			})
		}
	}

	t.Run("check later columns", func(t *testing.T) {
		assertCollationAudit(t, "ALTER TABLE arch ADD COLUMN name VARCHAR(50) COLLATE utf8mb4_bin, ADD COLUMN address VARCHAR(50) COLLATE utf8mb4_unicode_ci", collationConfig(), "列`address`指定的排序规则`utf8mb4_unicode_ci`不符合要求")
	})
	t.Run("case insensitive", func(t *testing.T) {
		cfg := collationConfig()
		cfg.TABLE_SUPPORT_COLLATION = []string{"UTF8MB4_BIN"}
		cfg.COLUMN_SUPPORT_COLLATION = []string{"UTF8MB4_BIN"}
		assertCollationAudit(t, "CREATE TABLE arch (address VARCHAR(50) COLLATE utf8mb4_bin) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin", cfg, "")
	})
	t.Run("collation does not bypass charset whitelist", func(t *testing.T) {
		cfg := collationConfig()
		cfg.COLUMN_SUPPORT_COLLATION = []string{"utf8_general_ci"}
		assertCollationAudit(t, "ALTER TABLE arch ADD COLUMN address VARCHAR(50) COLLATE utf8_general_ci", cfg, "列`address`指定的字符集`utf8`不符合要求")
	})
	t.Run("matching charset still required", func(t *testing.T) {
		cfg := collationConfig()
		cfg.COLUMN_SUPPORT_CHARSET = []string{"utf8", "utf8mb4"}
		assertCollationAudit(t, "ALTER TABLE arch ADD COLUMN address VARCHAR(50) CHARACTER SET utf8 COLLATE utf8mb4_bin", cfg, "应指定前缀为utf8_的排序规则")
	})
	t.Run("disabled checks", func(t *testing.T) {
		cfg := collationConfig()
		cfg.CHECK_TABLE_CHARSET = false
		cfg.CHECK_COLUMN_CHARSET = false
		assertCollationAudit(t, "CREATE TABLE arch (address VARCHAR(50) COLLATE utf8mb4_unicode_ci) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci", cfg, "")
	})
	for _, allowed := range [][]string{nil, {}} {
		t.Run(fmt.Sprintf("legacy whitelist %v", allowed), func(t *testing.T) {
			cfg := collationConfig()
			cfg.TABLE_SUPPORT_COLLATION = allowed
			cfg.COLUMN_SUPPORT_COLLATION = allowed
			assertCollationAudit(t, "CREATE TABLE arch (address VARCHAR(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci", cfg, "")
		})
	}
}

func collationConfig() *config.AuditConfiguration {
	return &config.AuditConfiguration{
		MAX_TABLE_NAME_LENGTH:    64,
		TABLE_SUPPORT_ENGINE:     []string{"InnoDB"},
		CHECK_TABLE_CHARSET:      true,
		TABLE_SUPPORT_CHARSET:    []map[string]string{{"charset": "utf8mb4", "recommend": "utf8mb4_general_ci"}},
		TABLE_SUPPORT_COLLATION:  []string{"utf8mb4_general_ci", "utf8mb4_bin"},
		CHECK_COLUMN_CHARSET:     true,
		COLUMN_SUPPORT_CHARSET:   []string{"utf8mb4"},
		COLUMN_SUPPORT_COLLATION: []string{"utf8mb4_general_ci", "utf8mb4_bin"},
	}
}

func assertCollationAudit(t *testing.T, sql string, cfg *config.AuditConfiguration, wantError string) {
	t.Helper()
	summary := collationAudit(t, sql, cfg)
	if wantError == "" {
		if len(summary) != 0 {
			t.Fatalf("unexpected audit errors: %v", summary)
		}
	} else if len(summary) != 1 || !strings.Contains(summary[0], wantError) {
		t.Fatalf("audit errors=%v, want one containing %q", summary, wantError)
	}
}
