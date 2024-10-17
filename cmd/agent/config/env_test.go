package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInitializeDefaultConfig(t *testing.T) {
	cases := []struct {
		name           string
		cliInput       []string
		envInput       map[string]string
		expected       *CliConfig
		expectedCliErr error
		expectedEnvErr error
	}{
		{
			name:     "ok",
			cliInput: []string{},
			envInput: map[string]string{},
			expected: &CliConfig{
				PollInterval:   DefaultPollInterval,
				ReportInterval: DefaultReportInterval,
				ServerURL:      DefaultServerURL,
				LogLevel:       DefaultLogLevel,
				HashKey:        DefaultHashKey,
				RateLimit:      DefaultRateLimit,
			},
			expectedCliErr: nil,
			expectedEnvErr: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := InitializeDefaultConfig()
			assert.Truef(t, compareConfigs(tc.expected, actual), "Expected %+v, but got %+v", tc.expected, actual)
		})
	}
}

func compareConfigs(expected, actual *CliConfig) bool {
	if expected.PollInterval != actual.PollInterval || expected.ReportInterval != actual.ReportInterval || expected.ServerURL != actual.ServerURL || expected.LogLevel != actual.LogLevel || expected.HashKey != actual.HashKey || expected.RateLimit != actual.RateLimit {
		return false
	}
	return true
}
