package query

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJsonParamsToArgs(t *testing.T) {
	tests := []struct {
		name    string
		raw     json.RawMessage
		order   []string
		want    []any
		wantErr bool
	}{
		{
			name:    "simple",
			raw:     json.RawMessage(`{"age": 30, "status": "active"}`),
			order:   []string{"age", "status"},
			want:    []any{float64(30), "active"}, // json unmarshal to any uses float64 for numbers
			wantErr: false,
		},
		{
			name:    "reordered",
			raw:     json.RawMessage(`{"age": 30, "status": "active"}`),
			order:   []string{"status", "age"},
			want:    []any{"active", float64(30)},
			wantErr: false,
		},
		{
			name:    "missing param",
			raw:     json.RawMessage(`{"age": 30}`),
			order:   []string{"age", "status"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := JsonParamsToArgs(tt.raw, tt.order)
			if (err != nil) != tt.wantErr {
				t.Errorf("JsonParamsToArgs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("JsonParamsToArgs() got = %v, want %v", got, tt.want)
			}
		})
	}
}
