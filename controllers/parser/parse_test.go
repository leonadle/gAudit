package parser

import (
	"testing"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

func TestParseTiDB81TTLAlterTable(t *testing.T) {
	sql := "ALTER TABLE bi_data.user_house_behavior_profile TTL = `updated_at` + INTERVAL 90 DAY TTL_ENABLE = 'ON' TTL_JOB_INTERVAL = '24h';"
	audit, warns, err := NewParse(sql, "", "")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(warns) != 0 {
		t.Fatalf("unexpected warnings: %v", warns)
	}
	if len(audit.TiStmt) != 1 {
		t.Fatalf("statement count = %d, want 1", len(audit.TiStmt))
	}
	stmt, ok := audit.TiStmt[0].(*ast.AlterTableStmt)
	if !ok {
		t.Fatalf("statement type = %T, want *ast.AlterTableStmt", audit.TiStmt[0])
	}
	if stmt.Table.Schema.O != "bi_data" || stmt.Table.Name.O != "user_house_behavior_profile" {
		t.Fatalf("table = %s.%s", stmt.Table.Schema.O, stmt.Table.Name.O)
	}
	if len(stmt.Specs) != 1 || stmt.Specs[0].Tp != ast.AlterTableOption {
		t.Fatalf("unexpected alter specs: %#v", stmt.Specs)
	}
	wantOptions := []ast.TableOptionType{
		ast.TableOptionTTL,
		ast.TableOptionTTLEnable,
		ast.TableOptionTTLJobInterval,
	}
	if len(stmt.Specs[0].Options) != len(wantOptions) {
		t.Fatalf("option count = %d, want %d", len(stmt.Specs[0].Options), len(wantOptions))
	}
	for i, want := range wantOptions {
		if got := stmt.Specs[0].Options[i].Tp; got != want {
			t.Fatalf("option[%d] = %v, want %v", i, got, want)
		}
	}
}

func TestParseTiDB81CommonBusinessDDL(t *testing.T) {
	tests := []string{
		"ALTER TABLE t TTL_ENABLE = 'OFF'",
		"ALTER TABLE t REMOVE TTL",
		"ALTER TABLE t AUTO_RANDOM_BASE = 100",
		"ALTER TABLE t SHARD_ROW_ID_BITS = 4",
		"ALTER TABLE t PLACEMENT POLICY = p1",
		"ALTER TABLE t CACHE",
		"ALTER TABLE t NOCACHE",
		"ALTER TABLE t SET TIFLASH REPLICA 1",
		"ALTER TABLE t ADD COLUMN IF NOT EXISTS c1 INT",
		"ALTER TABLE t DROP COLUMN IF EXISTS c1",
		"ALTER TABLE t RENAME COLUMN c1 TO c2",
		"ALTER TABLE t ALTER INDEX idx_c1 INVISIBLE",
		"ALTER TABLE t ADD PARTITION (PARTITION p1 VALUES LESS THAN (100))",
		"ALTER TABLE t TRUNCATE PARTITION p1",
		"ALTER TABLE t REORGANIZE PARTITION p0 INTO (PARTITION p0 VALUES LESS THAN (50), PARTITION p1 VALUES LESS THAN (100))",
		"ALTER TABLE t EXCHANGE PARTITION p1 WITH TABLE t_exchange",
		"CREATE INDEX IF NOT EXISTS idx_c1 ON t (c1)",
		"CREATE UNIQUE INDEX IF NOT EXISTS uniq_c1 ON t (c1)",
		"CREATE INDEX idx_expr ON t ((c1 + c2))",
		"DROP INDEX IF EXISTS idx_c1 ON t",
	}
	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			_, warns, err := NewParse(sql, "", "")
			if err != nil {
				t.Fatalf("parse failed: %v", err)
			}
			if len(warns) != 0 {
				t.Fatalf("unexpected warnings: %v", warns)
			}
		})
	}
}

func TestParseTiDBTTLRejectsInvalidOptions(t *testing.T) {
	tests := []string{
		"ALTER TABLE t TTL_ENABLE = 'INVALID'",
		"ALTER TABLE t TTL_JOB_INTERVAL = '@monthly'",
	}
	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			_, _, err := NewParse(sql, "", "")
			if err == nil {
				t.Fatal("expected parse error")
			}
		})
	}
}
