/*
@Time    :   2022/06/24 10:18:49
@Author  :   zongfei.fu
@Desc    :   获取目标数据库元信息
*/

package inspect

import (
	"fmt"
	"sqlSyntaxAudit/common/kv"
	"sqlSyntaxAudit/common/utils"
	"strings"

	mysqlapi "github.com/go-sql-driver/mysql"
)

// ShowCreateTable
func ShowCreateTable(table string, db *utils.DB, kv *kv.KVCache) (data interface{}, err error) {
	// 返回表结构
	data = kv.Get(table)
	if data != nil {
		return data, nil
	}
	query := fmt.Sprintf("show create table `%s`", table)
	result, err := db.FetchRows(query)
	if err != nil {
		return nil, err
	}
	var createStatement string
	for _, sql := range *result {
		createStatement = sql["Create Table"].(string)
	}

	var warns []error
	data, warns, err = NewParse(createStatement, "", "")
	if len(warns) > 0 {
		return nil, fmt.Errorf("Parse Warning: %s", utils.ErrsJoin("; ", warns))
	}
	if err != nil {
		return nil, fmt.Errorf("sql解析错误:%s", err.Error())
	}
	kv.Put(table, data)
	return data, nil
}

// descTable
func DescTable(table string, db *utils.DB) (error, string) {
	// 检查表是否存在
	err := db.Exec(fmt.Sprintf("desc `%s`", table))
	if me, ok := err.(*mysqlapi.MySQLError); ok {
		if me.Number == 1146 {
			// 表不存在
			return err, fmt.Sprintf("表`%s`不存在", table)
		} else if me.Number == 1045 {
			return err, fmt.Sprintf("访问目标数据库%s:%d失败,%s", db.Host, db.Port, err.Error())
		}
	}
	return nil, fmt.Sprintf("表`%s`已经存在", table)
}

// 获取DB变量
func GetDBVars(db *utils.DB) (map[string]string, error) {
	result, err := db.FetchRows("show variables where Variable_name in  ('innodb_large_prefix','version', 'character_set_database')")
	if err != nil {
		return nil, err
	}
	data := make(map[string]string)
	for _, row := range *result {
		if row["Variable_name"] == "version" {
			data["dbVersion"] = row["Value"].(string)
		}
		if row["Variable_name"] == "character_set_database" {
			data["dbCharset"] = row["Value"].(string)
		}
		if row["Variable_name"] == "innodb_large_prefix" {
			var largePrefix string
			switch row["Value"].(string) {
			case "0":
				largePrefix = "OFF"
			case "1":
				largePrefix = "ON"
			default:
				largePrefix = strings.ToUpper(row["Value"].(string))
			}
			data["largePrefix"] = largePrefix
		}
	}
	return data, nil
}
