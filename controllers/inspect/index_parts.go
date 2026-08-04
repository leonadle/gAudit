package inspect

import (
	"strings"

	"github.com/pingcap/tidb/pkg/parser/ast"
	"github.com/pingcap/tidb/pkg/parser/format"
)

// indexPartIdentity returns a stable identity for a column or expression index
// part. The boolean reports whether the part directly references a column.
func indexPartIdentity(part *ast.IndexPartSpecification) (string, bool) {
	if part == nil {
		return "", false
	}
	if part.Column != nil {
		return part.Column.Name.L, true
	}
	if part.Expr == nil {
		return "", false
	}
	var restored strings.Builder
	ctx := format.NewRestoreCtx(format.DefaultRestoreFlags, &restored)
	if err := part.Expr.Restore(ctx); err != nil {
		return "", false
	}
	return strings.ToLower(restored.String()), false
}
