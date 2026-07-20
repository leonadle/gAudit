package process

import (
	"sqlSyntaxAudit/config"
	"strings"
	"testing"

	"github.com/pingcap/tidb/parser/mysql"
)

func TestCheckColumnDefaultValueDate(t *testing.T) {
	tests := []struct {
		name         string
		defaultValue interface{}
		wantErr      bool
	}{
		{name: "valid date", defaultValue: "2026-07-20"},
		{name: "empty string", defaultValue: "", wantErr: true},
		{name: "zero date", defaultValue: "0000-00-00", wantErr: true},
		{name: "invalid calendar date", defaultValue: "2026-02-29", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := ColOptions{
				Table:           "orders",
				Column:          "target_trans_date",
				Tp:              mysql.TypeDate,
				HasDefaultValue: true,
				DefaultValue:    tt.defaultValue,
				AuditConfig:     &config.AuditConfiguration{},
			}

			err := col.CheckColumnDefaultValue()
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "DATE类型默认值不合法") {
					t.Fatalf("CheckColumnDefaultValue() error = %v, want invalid date default error", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CheckColumnDefaultValue() error = %v, want nil", err)
			}
		})
	}
}
