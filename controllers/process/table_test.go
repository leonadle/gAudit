package process

import (
	"sqlSyntaxAudit/config"
	"testing"
)

func tableAuditConfig() *config.AuditConfiguration {
	return &config.AuditConfiguration{
		MAX_TABLE_NAME_LENGTH:   32,
		CHECK_IDENTIFIER:        true,
		CHECK_IDENTIFER_KEYWORD: true,
	}
}

func TestTableOptionsAllowQualifiedTableName(t *testing.T) {
	opts := &TableOptions{
		Table:       "test.ios_refund",
		AuditConfig: tableAuditConfig(),
	}

	if err := opts.CheckTableLength(); err != nil {
		t.Fatalf("qualified table name should use the table part for length checks: %v", err)
	}
	if err := opts.CheckTableIdentifer(); err != nil {
		t.Fatalf("qualified table name should pass identifier checks: %v", err)
	}
	if err := opts.CheckTableIdentiferKeyword(); err != nil {
		t.Fatalf("qualified table name should pass keyword checks: %v", err)
	}
}

func TestTableOptionsRejectInvalidQualifiedIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		table string
	}{
		{name: "invalid schema", table: "test-db.ios_refund"},
		{name: "invalid table", table: "test.ios-refund"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &TableOptions{Table: tt.table, AuditConfig: tableAuditConfig()}
			if err := opts.CheckTableIdentifer(); err == nil {
				t.Fatalf("expected identifier validation error for %q", tt.table)
			}
		})
	}
}

func TestTableOptionsRejectKeywordInQualifiedIdentifier(t *testing.T) {
	opts := &TableOptions{
		Table:       "select.ios_refund",
		AuditConfig: tableAuditConfig(),
	}

	if err := opts.CheckTableIdentiferKeyword(); err == nil {
		t.Fatal("expected keyword validation error for schema")
	}
}
