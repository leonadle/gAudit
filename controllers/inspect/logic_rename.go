/*
@Time    :   2022/08/25 16:42:48
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/common/utils"
)

// LogicRenameTable
func LogicRenameTable(v *TraverseRenameTable, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if !r.AuditConfig.ENABLE_RENAME_TABLE_NAME {
		r.Summary = append(r.Summary, "不允许执行RENAME TABLE操作")
		return
	}
	// 禁止审核指定的表
	if len(r.AuditConfig.DISABLE_AUDIT_DDL_TABLES) > 0 {
		for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
			for _, t := range v.tables {
				schema, tableName := utils.SplitTableName(t.OldTable)
				if schema == "" {
					schema = r.DB.Database
				}
				if item.DB == schema && utils.IsContain(item.Tables, tableName) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核，原因: %s", schema, tableName, item.Reason))
				}
			}
		}
	}
	var oldTables []string
	// 旧表必须存在
	for _, t := range v.tables {
		if err, msg := DescTable(t.OldTable, r.DB, r.AuditConfig); err != nil {
			r.Summary = append(r.Summary, msg)
		} else {
			oldTables = append(oldTables, t.OldTable)
		}
	}
	// 新表不能存在
	for _, t := range v.tables {
		// 支持语法 rename table test to test_old, test_new to test
		if len(oldTables) > 0 && utils.IsContain(oldTables, t.NewTable) {
			continue
		}
		if err := checkCrossDBAudit(t.NewTable, r.DB, r.AuditConfig); err != nil {
			r.Summary = append(r.Summary, err.Error())
			continue
		}
		if err, msg := DescTable(t.NewTable, r.DB, r.AuditConfig); err == nil {
			r.Summary = append(r.Summary, msg)
		}
	}
}
