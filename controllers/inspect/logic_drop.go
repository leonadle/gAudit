/*
@Time    :   2022/07/06 10:12:27
@Author  :   zongfei.fu
@Desc    :   None
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/common/utils"
)

// LogicDropTable
func LogicDropTable(v *TraverseDropTable, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.IsHasDropTable {
		if !r.AuditConfig.ENABLE_DROP_TABLE {
			r.Summary = append(r.Summary, fmt.Sprintf("禁止DROP[表%s]", v.Tables))
			return
		}
		// 禁止审核指定的表
		if len(r.AuditConfig.DISABLE_AUDIT_DDL_TABLES) > 0 {
			for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
				for _, table := range v.Tables {
					schema, tableName := utils.SplitTableName(table)
					if schema == "" {
						schema = r.DB.Database
					}
					if item.DB == schema && utils.IsContain(item.Tables, tableName) {
						r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核，原因: %s", schema, tableName, item.Reason))
					}
				}
			}
		}
		// 检查表是否存在
		for _, table := range v.Tables {
			if err, msg := DescTable(table, r.DB, r.AuditConfig); err != nil {
				r.Summary = append(r.Summary, msg)
			}
		}
	}
}

// LogicTruncateTable
func LogicTruncateTable(v *TraverseTruncateTable, r *Rule) {
	if v.IsMatch == 0 {
		return
	}
	if v.IsHasTruncateTable {
		if !r.AuditConfig.ENABLE_TRUNCATE_TABLE {
			r.Summary = append(r.Summary, fmt.Sprintf("禁止TRUNCATE[表%s]", v.Table))
			return
		}
		// 禁止审核指定的表
		if len(r.AuditConfig.DISABLE_AUDIT_DDL_TABLES) > 0 {
			for _, item := range r.AuditConfig.DISABLE_AUDIT_DDL_TABLES {
				schema, tableName := utils.SplitTableName(v.Table)
				if schema == "" {
					schema = r.DB.Database
				}
				if item.DB == schema && utils.IsContain(item.Tables, tableName) {
					r.Summary = append(r.Summary, fmt.Sprintf("表`%s`.`%s`被限制进行DDL语法审核，原因: %s", schema, tableName, item.Reason))
				}
			}
		}
		// 检查表是否存在
		if err, msg := DescTable(v.Table, r.DB, r.AuditConfig); err != nil {
			r.Summary = append(r.Summary, msg)
		}
	}
}
