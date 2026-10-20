package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCollationConfigCompatibility(t *testing.T) {
	for _, tt := range []struct {
		name       string
		json       string
		wantLength int
	}{
		{"omitted", `{}`, 0},
		{"empty", `{"TABLE_SUPPORT_COLLATION":[],"COLUMN_SUPPORT_COLLATION":[]}`, 0},
		{"configured", `{"TABLE_SUPPORT_COLLATION":["utf8mb4_general_ci","utf8mb4_bin"],"COLUMN_SUPPORT_COLLATION":["utf8mb4_general_ci","utf8mb4_bin"]}`, 2},
	} {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "config.json")
			if err := os.WriteFile(path, []byte(tt.json), 0600); err != nil {
				t.Fatal(err)
			}
			cfg := InitializeAuditConfig(path)
			if len(cfg.TABLE_SUPPORT_COLLATION) != tt.wantLength || len(cfg.COLUMN_SUPPORT_COLLATION) != tt.wantLength {
				t.Fatalf("unexpected collation configuration: table=%v, column=%v", cfg.TABLE_SUPPORT_COLLATION, cfg.COLUMN_SUPPORT_COLLATION)
			}
		})
	}
}
