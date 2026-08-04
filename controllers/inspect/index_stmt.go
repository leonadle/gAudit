package inspect

import (
	"fmt"
	"sqlSyntaxAudit/common/kv"

	"github.com/pingcap/tidb/pkg/parser/ast"
)

// IndexStmt adapts standalone CREATE/DROP INDEX statements to the existing
// ALTER TABLE index audit rules while preserving the original SQL in output.
func (c *Checker) IndexStmt(stmt ast.StmtNode, kv *kv.KVCache, fingerID string) (ReturnData, string) {
	originalSQL := stmt.Text()
	alterStmt, err := indexStmtToAlter(stmt)
	if err != nil {
		return ReturnData{
			FingerId: fingerID,
			Query:    originalSQL,
			Type:     "DDL",
			Level:    "WARN",
			Summary:  []string{err.Error()},
		}, ""
	}
	var synthetic ast.StmtNode = alterStmt
	data, mergeAlter := c.AlterTableStmt(synthetic, kv, fingerID)
	data.Query = originalSQL
	return data, mergeAlter
}

func indexStmtToAlter(stmt ast.StmtNode) (*ast.AlterTableStmt, error) {
	switch indexStmt := stmt.(type) {
	case *ast.CreateIndexStmt:
		constraintType := ast.ConstraintIndex
		switch indexStmt.KeyType {
		case ast.IndexKeyTypeNone:
		case ast.IndexKeyTypeUnique:
			constraintType = ast.ConstraintUniqIndex
		case ast.IndexKeyTypeFullText:
			constraintType = ast.ConstraintFulltext
		case ast.IndexKeyTypeSpatial:
			return nil, fmt.Errorf("TiDB不支持SPATIAL索引`%s`", indexStmt.IndexName)
		}
		return &ast.AlterTableStmt{
			Table: indexStmt.Table,
			Specs: []*ast.AlterTableSpec{{
				Tp: ast.AlterTableAddConstraint,
				Constraint: &ast.Constraint{
					IfNotExists: indexStmt.IfNotExists,
					Tp:          constraintType,
					Name:        indexStmt.IndexName,
					Keys:        indexStmt.IndexPartSpecifications,
					Option:      indexStmt.IndexOption,
				},
			}},
		}, nil
	case *ast.DropIndexStmt:
		return &ast.AlterTableStmt{
			Table: indexStmt.Table,
			Specs: []*ast.AlterTableSpec{{
				Tp:       ast.AlterTableDropIndex,
				Name:     indexStmt.IndexName,
				IfExists: indexStmt.IfExists,
			}},
		}, nil
	default:
		return nil, fmt.Errorf("不支持的索引语句类型%T", stmt)
	}
}
