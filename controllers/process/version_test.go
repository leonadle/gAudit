package process

import "testing"

func TestDbVersionTiDBInt(t *testing.T) {
	tests := []struct {
		version string
		want    int
	}{
		{"5.7.25-TiDB-v5.0.4", 50004},
		{"8.0.11-TiDB-v8.1.2", 80102},
		{"8.0.11-TIDB-8.1.0", 80100},
		{"8.0.36", 0},
		{"unknown-TiDB", 0},
	}
	for _, tt := range tests {
		t.Run(tt.version, func(t *testing.T) {
			got := (&DbVersion{Version: tt.version}).TiDBInt()
			if got != tt.want {
				t.Fatalf("TiDBInt() = %d, want %d", got, tt.want)
			}
		})
	}
}
