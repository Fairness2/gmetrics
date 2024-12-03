// File: parse_test.go
package config

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
				if setErr := os.Setenv(k, v); setErr != nil {
					assert.Error(t, setErr, "os.Setenv(%q, %q) returned unexpected error", k, v)
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
				if setErr := os.Setenv(k, v); setErr != nil {
					assert.Error(t, setErr, "os.Setenv(%q, %q) returned unexpected error", k, v)
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
    "restore": true,
    "store_interval": "1s",
    "store_file": "/path/to/file.db",
    "database_dsn": "123",
    "crypto_key": "/path/to/key.pem"
}`,
			getCnf: func(t *testing.T) *CliConfig {
				return &CliConfig{
					Address:        DefaultServerURL,
					StoreInterval:  DefaultStoreInterval,
					Restore:        DefaultRestore,
					FileStorage:    DefaultFilePath,
					DatabaseDSN:    DefaultDatabaseDSN,
					CryptoKeyPath:  "",
					ConfigFilePath: testFilePath,
				}
			},
			want: &CliConfig{
				Address:        "localhost:8086",
				StoreInterval:  1,
				Restore:        true,
				FileStorage:    "/path/to/file.db",
				DatabaseDSN:    "123",
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
    "restore": false,
    "store_interval": "2s",
    "store_file": "/path/to/file1.db",
    "database_dsn": "1243",
    "crypto_key": "/path/to/key1.pem"
}`,
			getCnf: func(t *testing.T) *CliConfig {
				return &CliConfig{
					Address:        "localhost:8085",
					StoreInterval:  1,
					Restore:        true,
					FileStorage:    "/path/to/file.db",
					DatabaseDSN:    "123",
					CryptoKeyPath:  "/path/to/key.pem",
					ConfigFilePath: testFilePath,
				}
			},
			want: &CliConfig{
				Address:        "localhost:8085",
				StoreInterval:  1,
				Restore:        true,
				FileStorage:    "/path/to/file.db",
				DatabaseDSN:    "123",
				CryptoKeyPath:  "/path/to/key.pem",
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

// Test cases for parseSubnet function
func TestParseSubnet(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "empty_subnet",
			in:      "",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "valid_subnet",
			in:      "192.0.2.0/24",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "invalid_subnet",
			in:      "192.0.2.0/42",
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSubnet(tt.in)
			if tt.wantNil {
				assert.Nil(t, got)
			} else {
				assert.NotNil(t, got)
			}
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
