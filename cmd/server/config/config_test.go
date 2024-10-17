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
				Address:       DefaultServerURL,
				LogLevel:      DefaultLogLevel,
				FileStorage:   DefaultFilePath,
				DatabaseDSN:   DefaultDatabaseDSN,
				StoreInterval: DefaultStoreInterval,
				Restore:       DefaultRestore,
				HashKey:       DefaultHashKey,
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
