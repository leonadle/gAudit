package utils

import (
	"sqlSyntaxAudit/controllers/parser"
	"testing"
)

func TestTableNameHelpers(t *testing.T) {
	if got := FormatTableName("", "arch"); got != "arch" {
		t.Fatalf("FormatTableName without schema = %q", got)
	}
	if got := FormatTableName("test02", "arch"); got != "test02.arch" {
		t.Fatalf("FormatTableName with schema = %q", got)
	}
	if got := QuoteTableName("test`02.my`test"); got != "`test``02`.`my``test`" {
		t.Fatalf("QuoteTableName escaped = %q", got)
	}
	if got := TableCacheKey("mytest", "test"); got != "test.mytest" {
		t.Fatalf("TableCacheKey unqualified = %q", got)
	}
	if got := TableCacheKey("test02.mytest", "test"); got != "test02.mytest" {
		t.Fatalf("TableCacheKey qualified = %q", got)
	}
	if !IsExplicitCrossDB("test02.mytest", "test") {
		t.Fatal("expected explicit cross-db table")
	}
	if IsExplicitCrossDB("test.mytest", "test") {
		t.Fatal("current schema should not be cross-db")
	}
}

func TestNormalizeCreateTableForParser(t *testing.T) {
	createStatement := "CREATE TABLE `house_source` (\n  `id` bigint NOT NULL,\n  `location` point DEFAULT NULL COMMENT '位置',\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB"
	got := NormalizeCreateTableForParser(createStatement)
	want := "CREATE TABLE `house_source` (\n  `id` bigint NOT NULL,\n  `location` varchar(255) DEFAULT NULL COMMENT '位置',\n  PRIMARY KEY (`id`)\n) ENGINE=InnoDB"
	if got != want {
		t.Fatalf("NormalizeCreateTableForParser() = %q, want %q", got, want)
	}
	if _, _, err := parser.NewParse(got, "", ""); err != nil {
		t.Fatalf("normalized create table should parse: %v", err)
	}
}
