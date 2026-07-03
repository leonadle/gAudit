package utils

import "strings"

// FormatTableName joins an optional schema and table name into the internal
// representation used by audit rules.
func FormatTableName(schema, table string) string {
	if schema == "" {
		return table
	}
	return schema + "." + table
}

// SplitTableName splits db.table into schema and table parts.
func SplitTableName(table string) (string, string) {
	parts := strings.SplitN(table, ".", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", table
}

func EffectiveTableSchema(table, currentDB string) string {
	schema, _ := SplitTableName(table)
	if schema != "" {
		return schema
	}
	return currentDB
}

func TableCacheKey(table, currentDB string) string {
	schema, name := SplitTableName(table)
	if schema == "" {
		schema = currentDB
	}
	return FormatTableName(schema, name)
}

func IsExplicitCrossDB(table, currentDB string) bool {
	schema, _ := SplitTableName(table)
	return schema != "" && !strings.EqualFold(schema, currentDB)
}

func QuoteIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}

func QuoteTableName(table string) string {
	schema, name := SplitTableName(table)
	if schema == "" {
		return QuoteIdentifier(name)
	}
	return QuoteIdentifier(schema) + "." + QuoteIdentifier(name)
}

func QuoteSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func DisplayTableName(table string) string {
	return QuoteTableName(table)
}
