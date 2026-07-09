package inspect

import (
	"sqlSyntaxAudit/controllers/parser"
	"testing"

	"github.com/pingcap/tidb/parser/ast"
)

func TestTableNameWithSchemaPreservesOriginalCase(t *testing.T) {
	audit, warns, err := parser.NewParse("insert into `t591`.`userMoneyNote` (`id`) values (1)", "", "")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(warns) > 0 {
		t.Fatalf("unexpected warnings: %v", warns)
	}
	stmt, ok := audit.TiStmt[0].(*ast.InsertStmt)
	if !ok {
		t.Fatalf("unexpected stmt type %T", audit.TiStmt[0])
	}
	v := &TraverseDMLInsertWithColumns{}
	v.CheckSelectItem(stmt.Table.TableRefs)
	if v.Table != "t591.userMoneyNote" {
		t.Fatalf("table name = %q, want %q", v.Table, "t591.userMoneyNote")
	}
}
