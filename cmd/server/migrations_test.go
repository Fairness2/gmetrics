package main

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestMigrations(t *testing.T) {
	tCases := []struct {
		desc       string
		shouldFail bool
	}{
		{
			desc:       "success",
			shouldFail: false,
		},
	}

	for _, tCase := range tCases {
		t.Run(tCase.desc, func(t *testing.T) {
			_, err := migrations()
			require.NoError(t, err)

			/*testDb, _ := sql.Open("sqlite3", ":memory:")
			err = m.Migrate(testDb)
			if tCase.shouldFail {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}*/
		})
	}
}
