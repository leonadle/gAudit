/*
@Time    :   2022/07/06 10:12:48
@Author  :   zongfei.fu
@Desc    :   None
*/

package process

import (
	"sqlSyntaxAudit/common/kv"
	"sqlSyntaxAudit/common/utils"
	"sqlSyntaxAudit/global"
	"strconv"
	"strings"

	"github.com/mitchellh/mapstructure"
)

// 对应explain的输出
type ExplainOutput struct {
	Table string `json:"Column:table"`
	// MySQL的Explain预估行数
	Rows int64 `json:"Column:rows"`

	// TiDB (v4.0及之后)的Explain预估行数存储在Count中
	EstRows interface{} `json:"Column:estRows"`
}

type Explain struct {
	DB  *utils.DB
	SQL string
	KV  *kv.KVCache
}

func (e Explain) ConvertToExplain() string {
	var explain []string
	explain = append(explain, "EXPLAIN ")
	explain = append(explain, e.SQL)
	return strings.Join(explain, "")
}

func (e *Explain) Get() (int64, error) {
	rows, err := e.DB.FetchRows(e.ConvertToExplain())
	if err != nil {
		return 0, err
	}
	// 赋值给结构体
	var data []ExplainOutput
	err = mapstructure.WeakDecode(rows, &data)
	if err != nil {
		return 0, err
	}
	// 获取db版本
	dbVersionIns := DbVersion{e.KV.Get("dbVersion").(string)}

	var AffectedRows []int64
	for _, item := range data {
		if dbVersionIns.IsTiDB() {
			// tidb的执行计划第一行可能是estRows=N/A
			floatEstRows, err := strconv.ParseFloat(item.EstRows.(string), 64)
			if err != nil {
				continue
			}
			// float64 -> int64
			int64EstRows := int64(floatEstRows)
			AffectedRows = append(AffectedRows, int64EstRows)
		} else {
			AffectedRows = append(AffectedRows, item.Rows)
		}
	}
	if global.App.AuditConfig.EXPLAIN_RULE == "first" {
		return AffectedRows[0], nil
	}
	if global.App.AuditConfig.EXPLAIN_RULE == "max" {
		return utils.MaxInt64(AffectedRows), nil
	}
	return 0, nil
}