package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFromCli(t *testing.T) {
	cases := []struct {
		name     string
		input    []string
		expected *CliConfig
	}{
		{
			name:  "empty",
			input: []string{},
			expected: &CliConfig{
				PollInterval:   DefaultPollInterval,
				ReportInterval: DefaultReportInterval,
				ServerURL:      "",
				LogLevel:       DefaultLogLevel,
				HashKey:        DefaultHashKey,
				RateLimit:      DefaultRateLimit,
			},
		},
		{
			name:  "flags_passed",
			input: []string{"-p=1", "-ll=loglevel", "-r=1", "-l=10", "-k=key", "-a=someAddress"},
			expected: &CliConfig{
				PollInterval:   1,
				LogLevel:       "loglevel",
				ServerURL:      "http://someAddress",
				ReportInterval: 1,
				RateLimit:      10,
				HashKey:        "key",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = append([]string{"cmd"}, tc.input...)
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
			actual := &CliConfig{}
			if err := parseFromCli(actual); err != nil {
				assert.Error(t, err, "parseFromCli(%q) returned unexpected error")
				return
			}
			assert.Truef(t, compareConfigs(tc.expected, actual), "Expected %+v, but got %+v", tc.expected, actual)
		})
	}
}

func TestParseFromEnv(t *testing.T) {
	cases := []struct {
		name     string
		input    map[string]string
		expected *CliConfig
	}{
		{
			name:  "empty",
			input: map[string]string{},
			expected: &CliConfig{
				PollInterval:   0,
				ReportInterval: 0,
				ServerURL:      "",
				LogLevel:       "",
				HashKey:        "",
				RateLimit:      0,
			},
		},
		{
			name: "environment_variables_set",
			input: map[string]string{
				"POLL_INTERVAL":   "1",
				"REPORT_INTERVAL": "1",
				"ADDRESS":         "someAddress",
				"LOG_LEVEL":       "loglevel",
				"RATE_LIMIT":      "10",
				"KEY":             "key",
			},
			expected: &CliConfig{
				PollInterval:   1,
				ReportInterval: 1,
				ServerURL:      "http://someAddress",
				LogLevel:       "loglevel",
				HashKey:        "key",
				RateLimit:      10,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tc.input {
				os.Setenv(k, v)
			}
			actual := &CliConfig{}
			if err := parseFromEnv(actual); err != nil {
				assert.Error(t, err, "parseFromEnv(%q) returned unexpected error")
				return
			}

			assert.Truef(t, compareConfigs(tc.expected, actual), "Expected %+v, but got %+v", tc.expected, actual)
		})
	}
}

func TestParse(t *testing.T) {
	cases := []struct {
		name           string
		cliInput       []string
		envInput       map[string]string
		expected       *CliConfig
		expectedCliErr error
		expectedEnvErr error
	}{
		{
			name:     "empty",
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
		{
			name:     "cli_flags_passed",
			cliInput: []string{"-p=1", "-ll=loglevel", "-r=1", "-l=10", "-k=key", "-a=someAddress"},
			envInput: map[string]string{},
			expected: &CliConfig{
				PollInterval:   1,
				LogLevel:       "loglevel",
				ServerURL:      "http://someAddress",
				ReportInterval: 1,
				RateLimit:      10,
				HashKey:        "key",
			},
			expectedCliErr: nil,
			expectedEnvErr: nil,
		},
		{
			name:     "env_variables_passed",
			cliInput: []string{},
			envInput: map[string]string{
				"POLL_INTERVAL":   "1",
				"REPORT_INTERVAL": "1",
				"ADDRESS":         "someAddress",
				"LOG_LEVEL":       "loglevel",
				"RATE_LIMIT":      "10",
				"KEY":             "key",
			},
			expected: &CliConfig{
				PollInterval:   1,
				ReportInterval: 1,
				ServerURL:      "http://someAddress",
				LogLevel:       "loglevel",
				HashKey:        "key",
				RateLimit:      10,
			},
			expectedCliErr: nil,
			expectedEnvErr: nil,
		},
		{
			name:     "env_has_more_priority_than_cli",
			cliInput: []string{"-p=21", "-ll=loglevel1", "-r=11", "-l=101", "-k=key1", "-a=someAddress1"},
			envInput: map[string]string{
				"POLL_INTERVAL":   "1",
				"REPORT_INTERVAL": "1",
				"ADDRESS":         "someAddress",
				"LOG_LEVEL":       "loglevel",
				"RATE_LIMIT":      "10",
				"KEY":             "key",
			},
			expected: &CliConfig{
				PollInterval:   1,
				ReportInterval: 1,
				ServerURL:      "http://someAddress",
				LogLevel:       "loglevel",
				HashKey:        "key",
				RateLimit:      10,
			},
			expectedCliErr: nil,
			expectedEnvErr: nil,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			os.Args = append([]string{"cmd"}, tc.cliInput...)
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)
			os.Clearenv()
			for k, v := range tc.envInput {
				os.Setenv(k, v)
			}
			actual, err := Parse()
			if err != nil {
				assert.Error(t, err, "Parse() returned unexpected error")
				return
			}
			assert.Truef(t, compareConfigs(tc.expected, actual), "Expected %+v, but got %+v", tc.expected, actual)
		})
	}
}
