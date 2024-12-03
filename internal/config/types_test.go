package config

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestDuration_UnmarshalJSON(t *testing.T) {
	type Some struct {
		D Duration `json:"d"`
	}
	tests := []struct {
		name    string
		jsonStr string
		want    Duration
		wantErr bool
	}{
		{
			name:    "float",
			jsonStr: `{"d":10000.0}`,
			want:    Duration{Duration: time.Duration(10000)},
			wantErr: false,
		},
		{
			name:    "int",
			jsonStr: `{"d":1}`,
			want:    Duration{Duration: time.Duration(1)},
			wantErr: false,
		},
		{
			name:    "duration_str",
			jsonStr: `{"d":"11s"}`,
			want:    Duration{Duration: time.Duration(11 * time.Second)},
			wantErr: false,
		},
		{
			name:    "invalid_str",
			jsonStr: `{"d":"invalid"}`,
			wantErr: true,
		},
		{
			name:    "invalid_type",
			jsonStr: `{"d":[]}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := Some{}
			err := json.Unmarshal([]byte(tt.jsonStr), &d)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, d.D)
		})
	}
}
