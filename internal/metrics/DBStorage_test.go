package metrics

import (
	"context"
	"database/sql"
	"testing"

	"github.com/golang/mock/gomock"
)

func BenchmarkDBStorage_SetMetrics(b *testing.B) {
	benchmarks := []struct {
		name     string
		gauges   map[string]Gauge
		counters map[string]Counter
	}{
		{
			name:     "one_value",
			gauges:   generateGaugesMap(1),
			counters: generateCounterMap(1),
		},
		{
			name:     "1000_value",
			gauges:   generateGaugesMap(1000),
			counters: generateCounterMap(1000),
		},
		{
			name:     "10000_value",
			gauges:   generateGaugesMap(1000),
			counters: generateCounterMap(1000),
		},
	}
	ctrl := gomock.NewController(b)
	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			ctx, cancel := context.WithCancel(context.TODO())
			defer cancel()
			executor := NewMockSQLExecutor(ctrl)
			executor.EXPECT().ExecContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&MockSQLResult{}, nil)
			executor.EXPECT().QueryRowContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&sql.Row{})
			executor.EXPECT().QueryContext(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&sql.Rows{}, nil)
			executor.EXPECT().BeginTx(gomock.Any(), gomock.Any()).AnyTimes().Return(&sql.Tx{}, nil)
			storage := NewMemStorage()
			store := &DBStorage{
				Storage:  storage,
				storeCtx: ctx,
				db:       executor,
				syncMode: false,
				close:    false,
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				store.SetGauges(bm.gauges)
				store.AddCounters(bm.counters)
			}
		})
	}
}

type MockSQLResult struct{}

func (m MockSQLResult) LastInsertId() (int64, error) {
	return 1, nil
}

func (m MockSQLResult) RowsAffected() (int64, error) {
	return 1, nil
}
