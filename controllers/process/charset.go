/*
@Time    :   2022/06/27 15:49:13
@Author  :   zongfei.fu
@Desc    :   None
*/

package process

import (
	"fmt"
	"sqlSyntaxAudit/common/utils"

	"github.com/pingcap/tidb/pkg/parser/charset"
	"github.com/pingcap/tidb/pkg/parser/mysql"
)

// 字符集
type TableCharset struct {
	Table   string
	Charset string
	Collate string
}
type ColumnCharset struct {
	Table           string
	Column          string
	Tp              byte
	Charset         string
	Collate         string
	CollateExplicit bool // 区分显式 COLLATE 与解析器为二进制类型填充的 binary
}
type Charset struct {
	SupportCharset     []string
	SupportCollation   []string
	RecommendCollation string
	Table              TableCharset
	Cols               []ColumnCharset
}

// 检查表的字符集
func (c *Charset) CheckTable() error {
	/*
		1.表必须指定字符集
		1.指定的字符集必须是可选的字符集
		2.指定排序规则的同时必须指定字符集
		3.检查字符集排序规则是否符合要求
	*/
	if !utils.IsContain(c.SupportCharset, c.Table.Charset) {
		if c.Table.Charset == "" {
			return fmt.Errorf("表`%s`必须指定字符集，可选字符集为%s【例如:DEFAULT CHARSET=utf8mb4】", c.Table.Table, c.SupportCharset)
		} else {
			return fmt.Errorf("表`%s`指定的字符集`%s`不符合要求，可选字符集为%s", c.Table.Table, c.Table.Charset, c.SupportCharset)
		}
	}

	if len(c.Table.Collate) > 0 && len(c.Table.Charset) > 0 && !utils.HasPrefix(c.Table.Collate, c.Table.Charset+"_", false) {
		// 检查排序规则的前缀，前缀必须为字符集+"_"
		return fmt.Errorf("表`%s`指定的字符集排序规则`%s`不符合要求，应指定前缀为%s的排序规则，推荐的字符集排序规则为%s", c.Table.Table, c.Table.Charset, c.Table.Charset+"_", c.RecommendCollation)
	}
	if c.Table.Collate != "" && len(c.SupportCollation) > 0 && !utils.IsContain(c.SupportCollation, c.Table.Collate) {
		return fmt.Errorf("表`%s`指定的排序规则`%s`不符合要求，可选排序规则为%s", c.Table.Table, c.Table.Collate, c.SupportCollation)
	}
	return nil
}

// 检查列字符集
func (c *Charset) CheckColumn() error {
	// https://dev.mysql.com/doc/refman/5.6/en/charset-column.html
	for _, col := range c.Cols {
		// 只限制显式 COLLATE；TEXT 同样受限，JSON/BLOB 等类型的隐式 binary 不受影响。
		if col.CollateExplicit && len(c.SupportCollation) > 0 && !utils.IsContain(c.SupportCollation, col.Collate) {
			return fmt.Errorf("列`%s`指定的排序规则`%s`不符合要求，可选排序规则为%s[表`%s`]", col.Column, col.Collate, c.SupportCollation, col.Table)
		}
		if col.Tp == mysql.TypeVarchar || col.Tp == mysql.TypeVarString || col.Tp == mysql.TypeString || col.Tp == mysql.TypeEnum || col.Tp == mysql.TypeSet {
			// 未启用排序规则白名单时保留原有配对要求。
			if len(c.SupportCollation) == 0 && (col.Charset == "" || col.Collate == "") {
				return fmt.Errorf("列`%s`必须同时指定字符集和排序规则[表`%s`]", col.Column, col.Table)
			}
			colCharset := col.Charset
			if colCharset == "" && col.Collate != "" {
				// 允许仅声明 COLLATE，字符集仍需满足列字符集白名单。
				collation, err := charset.GetCollationByName(col.Collate)
				if err != nil {
					return fmt.Errorf("列`%s`指定的排序规则`%s`无效[表`%s`]", col.Column, col.Collate, col.Table)
				}
				colCharset = collation.CharsetName
			}
			if colCharset != "" && len(c.SupportCharset) > 0 && !utils.IsContain(c.SupportCharset, colCharset) {
				return fmt.Errorf("列`%s`指定的字符集`%s`不符合要求，可选字符集为%s[表`%s`]", col.Column, colCharset, c.SupportCharset, col.Table)
			}
			// 检查排序规则
			if colCharset != "" && col.Collate != "" && !utils.HasPrefix(col.Collate, colCharset+"_", false) {
				// 检查排序规则的前缀，前缀必须为字符集+"_"
				return fmt.Errorf("列`%s`指定的字符集排序规则`%s`不符合要求，应指定前缀为%s的排序规则[表`%s`]", col.Column, col.Collate, colCharset+"_", col.Table)
			}
		}
	}

	return nil
}
