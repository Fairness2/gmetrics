package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
				Address:       DefaultServerURL,
				LogLevel:      DefaultLogLevel,
				FileStorage:   DefaultFilePath,
				DatabaseDSN:   DefaultDatabaseDSN,
				StoreInterval: DefaultStoreInterval,
				Restore:       DefaultRestore,
				HashKey:       DefaultHashKey,
			},
		},
		{
			name:  "flags_passed",
			input: []string{"-a=someAddress", "-ll=loglevel", "-f=path", "-d=dsn", "-i=1", "-r=true", "-k=key"},
			expected: &CliConfig{
				Address:       "someAddress",
				LogLevel:      "loglevel",
				FileStorage:   "path",
				DatabaseDSN:   "dsn",
				StoreInterval: 1,
				Restore:       true,
				HashKey:       "key",
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

func compareConfigs(expected, actual *CliConfig) bool {
	if expected.Address != actual.Address ||
		expected.LogLevel != actual.LogLevel ||
		expected.FileStorage != actual.FileStorage ||
		expected.DatabaseDSN != actual.DatabaseDSN ||
		expected.StoreInterval != actual.StoreInterval ||
		expected.Restore != actual.Restore ||
		expected.HashKey != actual.HashKey {
		return false
	}
	return true
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
				Address:       "",
				LogLevel:      "",
				FileStorage:   "",
				DatabaseDSN:   "",
				StoreInterval: 0,
				Restore:       false,
				HashKey:       "",
			},
		},
		{
			name: "environment_variables_set",
			input: map[string]string{
				"ADDRESS":        "someAddress",
				"LOG_LEVEL":      "loglevel",
				"FILE_STORAGE":   "path",
				"DATABASE_DSN":   "dsn",
				"STORE_INTERVAL": "1",
				"RESTORE":        "true",
				"KEY":            "key",
			},
			expected: &CliConfig{
				Address:       "someAddress",
				LogLevel:      "loglevel",
				FileStorage:   "path",
				DatabaseDSN:   "dsn",
				StoreInterval: 1,
				Restore:       true,
				HashKey:       "key",
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
				Address:       DefaultServerURL,
				LogLevel:      DefaultLogLevel,
				FileStorage:   DefaultFilePath,
				DatabaseDSN:   DefaultDatabaseDSN,
				StoreInterval: 0,
				Restore:       DefaultRestore,
				HashKey:       DefaultHashKey,
			},
			expectedCliErr: nil,
			expectedEnvErr: nil,
		},
		{
			name:     "cli_flags_passed",
			cliInput: []string{"-a=someAddress", "-ll=loglevel", "-f=path", "-d=dsn", "-i=1", "-r=true", "-k=key"},
			envInput: map[string]string{"STORE_INTERVAL": "-1"},
			expected: &CliConfig{
				Address:       "someAddress",
				LogLevel:      "loglevel",
				FileStorage:   "path",
				DatabaseDSN:   "dsn",
				StoreInterval: 1,
				Restore:       true,
				HashKey:       "key",
			},
			expectedCliErr: nil,
			expectedEnvErr: nil,
		},
		{
			name:     "env_variables_passed",
			cliInput: []string{},
			envInput: map[string]string{
				"ADDRESS":        "someAddress",
				"LOG_LEVEL":      "loglevel",
				"FILE_STORAGE":   "path",
				"DATABASE_DSN":   "dsn",
				"STORE_INTERVAL": "1",
				"RESTORE":        "true",
				"KEY":            "key",
			},
			expected: &CliConfig{
				Address:       "someAddress",
				LogLevel:      "loglevel",
				FileStorage:   "path",
				DatabaseDSN:   "dsn",
				StoreInterval: 1,
				Restore:       true,
				HashKey:       "key",
			},
			expectedCliErr: nil,
			expectedEnvErr: nil,
		},
		{
			name:     "env_has_more_priority_than_cli",
			cliInput: []string{"-a=someAddress1", "-ll=loglevel1", "-f=path1", "-d=dsn1", "-i=2", "-r=false", "-k=key1"},
			envInput: map[string]string{
				"ADDRESS":        "someAddress",
				"LOG_LEVEL":      "loglevel",
				"FILE_STORAGE":   "path",
				"DATABASE_DSN":   "dsn",
				"STORE_INTERVAL": "1",
				"RESTORE":        "true",
				"KEY":            "key",
			},
			expected: &CliConfig{
				Address:       "someAddress",
				LogLevel:      "loglevel",
				FileStorage:   "path",
				DatabaseDSN:   "dsn",
				StoreInterval: 1,
				Restore:       true,
				HashKey:       "key",
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
