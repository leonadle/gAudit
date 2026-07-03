package utils

import "testing"

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
