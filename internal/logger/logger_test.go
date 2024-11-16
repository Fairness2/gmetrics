package logger

import (
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		name      string
		level     string
		wantLevel string
		wantErr   bool
	}{
		{
			name:      "debug_level",
			level:     "debug",
			wantLevel: "debug",
			wantErr:   false,
		},
		{
			name:      "info_level",
			level:     "info",
			wantLevel: "info",
			wantErr:   false,
		},
		{
			name:      "non_existing_level",
			level:     "non-existent-level",
			wantLevel: "info",
			wantErr:   true,
		},
		{
			name:      "warn_level",
			level:     "warn",
			wantLevel: "warn",
			wantErr:   false,
		},
		{
			name:      "empty_level",
			level:     "",
			wantLevel: "info",
			wantErr:   false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			level, err := ParseLevel(tc.level)
			if tc.wantErr {
				assert.Error(t, err, "New() error = %v, wantErr %v", err, tc.wantErr)
			} else {
				assert.NoError(t, err, "New() error = %v, wantErr %v", err, tc.wantErr)
				assert.Equal(t, tc.wantLevel, level.String(), "New() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestNew(t *testing.T) {
	cases := []struct {
		name        string
		level       zap.AtomicLevel
		expectError bool
	}{
		{
			name:        "test_with_debug_level",
			level:       zap.NewAtomicLevelAt(zap.DebugLevel),
			expectError: false,
		},
		{
			name:        "test_with_error_level",
			level:       zap.AtomicLevel{},
			expectError: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.level)
			if tc.expectError {
				assert.Error(t, err, "New() error = %v, wantErr %v", err, tc.expectError)
			} else {
				assert.NoError(t, err, "New() error = %v, wantErr %v", err, tc.expectError)
			}
		})
	}
}

func TestInitLogger(t *testing.T) {
	cases := []struct {
		name        string
		level       string
		expectError bool
	}{
		{
			name:        "test_with_debug_level",
			level:       "debug",
			expectError: false,
		},
		{
			name:        "non-existent-level",
			level:       "non-existent-level",
			expectError: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			_, err := InitLogger(tc.level)
			if tc.expectError {
				assert.Error(t, err, "Initlogger() error = %v, wantErr %v", err, tc.expectError)
			} else {
				assert.NoError(t, err, "Initlogger() error = %v, wantErr %v", err, tc.expectError)
			}
		})
	}
}
