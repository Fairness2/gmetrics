package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	testFilePath    = "test_config.json"
	nonExistentFile = "non_existent_file.json"
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
				if setErr := os.Setenv(k, v); setErr != nil {
					assert.Error(t, setErr, "Failed to set environment variable %q", k)
					return
				}
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
				if setErr := os.Setenv(k, v); setErr != nil {
					assert.Error(t, setErr, "Failed to set environment variable %q", k)
					return
				}
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

func createFileWithContent(path string, content []byte) {
	os.WriteFile(path, content, os.ModePerm)
}

func TestParseFromFile(t *testing.T) {
	tests := []struct {
		name       string
		cfgPath    string
		want       *CliConfig
		wantErr    bool
		getCnf     func(t *testing.T) *CliConfig
		fileConfig string
	}{
		{
			name:    "no_config_file_provided",
			cfgPath: "",
			want:    &CliConfig{},
			getCnf: func(t *testing.T) *CliConfig {
				return &CliConfig{
					ConfigFilePath: "",
				}
			},
			wantErr: false,
		},
		{
			name:    "config_file_does_not_exist",
			cfgPath: nonExistentFile,
			want:    &CliConfig{},
			getCnf: func(t *testing.T) *CliConfig {
				return &CliConfig{
					ConfigFilePath: nonExistentFile,
				}
			},
			wantErr: true,
		},
		{
			name:    "config_file_exists_with_valid_data",
			cfgPath: testFilePath,
			fileConfig: `{
    "address": "localhost:8086",
    "report_interval": "1s",
    "poll_interval": "1s",
    "crypto_key": "/path/to/key.pem"
}`,
			getCnf: func(t *testing.T) *CliConfig {
				return &CliConfig{
					ServerURL:      DefaultServerURL,
					ReportInterval: DefaultReportInterval,
					PollInterval:   DefaultPollInterval,
					CryptoKeyPath:  "",
					ConfigFilePath: testFilePath,
				}
			},
			want: &CliConfig{
				ServerURL:      "http://localhost:8086",
				ReportInterval: 1,
				PollInterval:   1,
				CryptoKeyPath:  "/path/to/key.pem",
				ConfigFilePath: testFilePath,
			},
			wantErr: false,
		},
		{
			name:    "config_file_exists_with_invalid_data",
			cfgPath: testFilePath,
			fileConfig: `{
    "address": "localhost:8`,
			getCnf: func(t *testing.T) *CliConfig {
				return &CliConfig{
					ConfigFilePath: testFilePath,
				}
			},
			want:    &CliConfig{},
			wantErr: true,
		},
		{
			name:    "config_file_exists_with_valid_data_but_config_not_default",
			cfgPath: testFilePath,
			fileConfig: `{
    "address": "localhost:8086",
    "report_interval": "1s",
    "poll_interval": "1s",
    "crypto_key": "/path/to/key.pem"
}`,
			getCnf: func(t *testing.T) *CliConfig {
				return &CliConfig{
					ServerURL:      "localhost:8086",
					ReportInterval: 2,
					PollInterval:   10,
					CryptoKeyPath:  "/path/to/key1.pem",
					ConfigFilePath: testFilePath,
				}
			},
			want: &CliConfig{
				ServerURL:      "localhost:8086",
				ReportInterval: 2,
				PollInterval:   10,
				CryptoKeyPath:  "/path/to/key1.pem",
				ConfigFilePath: testFilePath,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfgPath == testFilePath {
				defer os.Remove(tt.cfgPath)
				createFileWithContent(tt.cfgPath, []byte(tt.fileConfig))
			}
			cnf := tt.getCnf(t)
			err := parseFromFile(cnf)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Truef(t, compareConfigs(tt.want, cnf), "Expected %+v, but got %+v", tt.want, cnf)
		})
	}
}
