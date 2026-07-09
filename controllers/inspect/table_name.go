package inspect

import (
	"sqlSyntaxAudit/common/utils"

	"github.com/pingcap/tidb/parser/ast"
)

func tableNameWithSchema(table *ast.TableName) string {
	if table == nil {
		return ""
	}
	return utils.FormatTableName(table.Schema.O, table.Name.O)
}
