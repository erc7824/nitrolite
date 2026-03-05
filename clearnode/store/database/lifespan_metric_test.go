package database

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetMetricID(t *testing.T) {
	t.Run("deterministic ID", func(t *testing.T) {
		id1, err := getMetricID("test_metric", "key1", "val1")
		require.NoError(t, err)

		id2, err := getMetricID("test_metric", "key1", "val1")
		require.NoError(t, err)

		assert.Equal(t, id1, id2)
	})

	t.Run("different labels produce different IDs", func(t *testing.T) {
		id1, err := getMetricID("test_metric", "key1", "val1")
		require.NoError(t, err)

		id2, err := getMetricID("test_metric", "key1", "val2")
		require.NoError(t, err)

		assert.NotEqual(t, id1, id2)
	})

	t.Run("different names produce different IDs", func(t *testing.T) {
		id1, err := getMetricID("metric_a")
		require.NoError(t, err)

		id2, err := getMetricID("metric_b")
		require.NoError(t, err)

		assert.NotEqual(t, id1, id2)
	})

	t.Run("no labels", func(t *testing.T) {
		id, err := getMetricID("metric_no_labels")
		require.NoError(t, err)
		assert.NotEmpty(t, id)
	})

	t.Run("ID starts with 0x", func(t *testing.T) {
		id, err := getMetricID("test")
		require.NoError(t, err)
		assert.Equal(t, "0x", id[:2])
	})
}

func TestRecordMetric(t *testing.T) {
	t.Run("insert new metric", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts := time.Now().Truncate(time.Second)
		err := store.RecordMetric("test_metric", decimal.NewFromInt(42), ts, "env", "prod")
		require.NoError(t, err)

		metric, err := store.GetLifetimeMetric("test_metric", "env", "prod")
		require.NoError(t, err)

		assert.Equal(t, "test_metric", metric.Name)
		assert.True(t, decimal.NewFromInt(42).Equal(metric.Value))
		assert.Equal(t, ts.UTC(), metric.LastTimestamp.UTC())
	})

	t.Run("upsert overwrites value and timestamp", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts1 := time.Now().Add(-time.Hour).Truncate(time.Second)
		err := store.RecordMetric("test_metric", decimal.NewFromInt(10), ts1, "env", "prod")
		require.NoError(t, err)

		ts2 := time.Now().Truncate(time.Second)
		err = store.RecordMetric("test_metric", decimal.NewFromInt(20), ts2, "env", "prod")
		require.NoError(t, err)

		metric, err := store.GetLifetimeMetric("test_metric", "env", "prod")
		require.NoError(t, err)

		assert.True(t, decimal.NewFromInt(20).Equal(metric.Value))
		assert.Equal(t, ts2.UTC(), metric.LastTimestamp.UTC())
	})

	t.Run("odd label count returns error", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		err := store.RecordMetric("test_metric", decimal.NewFromInt(1), time.Now(), "key_only")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "labels must be key-value pairs")
	})

	t.Run("no labels", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts := time.Now().Truncate(time.Second)
		err := store.RecordMetric("simple_metric", decimal.NewFromInt(100), ts)
		require.NoError(t, err)

		metric, err := store.GetLifetimeMetric("simple_metric")
		require.NoError(t, err)

		assert.True(t, decimal.NewFromInt(100).Equal(metric.Value))
	})

	t.Run("stores labels as JSON map", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts := time.Now().Truncate(time.Second)
		err := store.RecordMetric("labeled_metric", decimal.NewFromInt(1), ts, "asset", "USDC", "status", "open")
		require.NoError(t, err)

		metric, err := store.GetLifetimeMetric("labeled_metric", "asset", "USDC", "status", "open")
		require.NoError(t, err)

		var labels map[string]string
		err = json.Unmarshal(metric.Labels, &labels)
		require.NoError(t, err)
		assert.Equal(t, "USDC", labels["asset"])
		assert.Equal(t, "open", labels["status"])
	})

	t.Run("multiple label pairs", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts := time.Now().Truncate(time.Second)
		err := store.RecordMetric("multi", decimal.NewFromInt(5), ts, "a", "1", "b", "2", "c", "3")
		require.NoError(t, err)

		metric, err := store.GetLifetimeMetric("multi", "a", "1", "b", "2", "c", "3")
		require.NoError(t, err)

		var labels map[string]string
		err = json.Unmarshal(metric.Labels, &labels)
		require.NoError(t, err)
		assert.Len(t, labels, 3)
		assert.Equal(t, "1", labels["a"])
		assert.Equal(t, "2", labels["b"])
		assert.Equal(t, "3", labels["c"])
	})

	t.Run("same name different labels are separate metrics", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts := time.Now().Truncate(time.Second)
		require.NoError(t, store.RecordMetric("m", decimal.NewFromInt(10), ts, "k", "a"))
		require.NoError(t, store.RecordMetric("m", decimal.NewFromInt(20), ts, "k", "b"))

		m1, err := store.GetLifetimeMetric("m", "k", "a")
		require.NoError(t, err)
		assert.True(t, decimal.NewFromInt(10).Equal(m1.Value))

		m2, err := store.GetLifetimeMetric("m", "k", "b")
		require.NoError(t, err)
		assert.True(t, decimal.NewFromInt(20).Equal(m2.Value))
	})
}

func TestGetLifetimeMetric(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		_, err := store.GetLifetimeMetric("nonexistent")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric not found")
	})

	t.Run("retrieves stored metric", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts := time.Now().Truncate(time.Second)
		require.NoError(t, store.RecordMetric("test", decimal.NewFromInt(99), ts, "env", "staging"))

		metric, err := store.GetLifetimeMetric("test", "env", "staging")
		require.NoError(t, err)

		assert.Equal(t, "test", metric.Name)
		assert.True(t, decimal.NewFromInt(99).Equal(metric.Value))
		assert.Equal(t, ts.UTC(), metric.LastTimestamp.UTC())
		assert.NotEmpty(t, metric.ID)
		assert.NotEmpty(t, metric.Labels)
	})

	t.Run("wrong labels returns not found", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts := time.Now().Truncate(time.Second)
		require.NoError(t, store.RecordMetric("test", decimal.NewFromInt(1), ts, "k", "v"))

		_, err := store.GetLifetimeMetric("test", "k", "wrong")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metric not found")
	})
}

func TestGetLifetimeMetricLastTimestamp(t *testing.T) {
	t.Run("no metrics returns zero time", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts, err := store.GetLifetimeMetricLastTimestamp("nonexistent")
		require.NoError(t, err)
		assert.True(t, ts.IsZero())
	})

	t.Run("returns most recent timestamp", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		ts1 := time.Now().Add(-2 * time.Hour).Truncate(time.Second)
		ts2 := time.Now().Add(-1 * time.Hour).Truncate(time.Second)
		ts3 := time.Now().Truncate(time.Second)

		require.NoError(t, store.RecordMetric("my_metric", decimal.NewFromInt(1), ts1, "label", "a"))
		require.NoError(t, store.RecordMetric("my_metric", decimal.NewFromInt(2), ts3, "label", "b"))
		require.NoError(t, store.RecordMetric("my_metric", decimal.NewFromInt(3), ts2, "label", "c"))

		latest, err := store.GetLifetimeMetricLastTimestamp("my_metric")
		require.NoError(t, err)
		assert.Equal(t, ts3.UTC(), latest.UTC())
	})

	t.Run("scoped to metric name", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		tsOld := time.Now().Add(-time.Hour).Truncate(time.Second)
		tsNew := time.Now().Truncate(time.Second)

		require.NoError(t, store.RecordMetric("metric_a", decimal.NewFromInt(1), tsOld))
		require.NoError(t, store.RecordMetric("metric_b", decimal.NewFromInt(1), tsNew))

		latest, err := store.GetLifetimeMetricLastTimestamp("metric_a")
		require.NoError(t, err)
		assert.Equal(t, tsOld.UTC(), latest.UTC())
	})
}

func TestCountActiveUsers(t *testing.T) {
	t.Run("no data returns only ALL with zero", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		results, err := store.CountActiveUsers(24 * time.Hour)
		require.NoError(t, err)

		require.Len(t, results, 1)
		assert.Equal(t, "ALL", results[0].Label)
		assert.Equal(t, uint64(0), results[0].Count)
	})

	t.Run("counts distinct users per asset", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		now := time.Now()
		db.Create(&UserBalance{UserWallet: "0xuser1", Asset: "USDC", Balance: decimal.NewFromInt(100), UpdatedAt: now})
		db.Create(&UserBalance{UserWallet: "0xuser2", Asset: "USDC", Balance: decimal.NewFromInt(200), UpdatedAt: now})
		db.Create(&UserBalance{UserWallet: "0xuser1", Asset: "ETH", Balance: decimal.NewFromInt(50), UpdatedAt: now})

		results, err := store.CountActiveUsers(24 * time.Hour)
		require.NoError(t, err)

		require.Len(t, results, 3)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(2), countByLabel["USDC"])
		assert.Equal(t, uint64(1), countByLabel["ETH"])
		assert.Equal(t, uint64(2), countByLabel["ALL"])
	})

	t.Run("respects time window", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		old := time.Now().Add(-48 * time.Hour)
		recent := time.Now()
		db.Create(&UserBalance{UserWallet: "0xold", Asset: "USDC", Balance: decimal.NewFromInt(100), UpdatedAt: old})
		db.Create(&UserBalance{UserWallet: "0xnew", Asset: "USDC", Balance: decimal.NewFromInt(200), UpdatedAt: recent})

		results, err := store.CountActiveUsers(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(1), countByLabel["USDC"])
		assert.Equal(t, uint64(1), countByLabel["ALL"])
	})
}

func TestCountActiveAppSessions(t *testing.T) {
	t.Run("no data returns empty", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)
		assert.Empty(t, results)
	})

	t.Run("counts sessions per application", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		now := time.Now()
		db.Create(&AppSessionV1{ID: "s1", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s2", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 2, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s3", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(2), countByLabel["app1"])
		assert.Equal(t, uint64(1), countByLabel["app2"])
	})

	t.Run("respects time window", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		old := time.Now().Add(-48 * time.Hour)
		recent := time.Now()
		db.Create(&AppSessionV1{ID: "s1", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: old})
		db.Create(&AppSessionV1{ID: "s2", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 2, UpdatedAt: recent})

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		assert.Equal(t, uint64(1), countByLabel["app1"])
	})

	t.Run("multiple applications with mixed statuses", func(t *testing.T) {
		db, cleanup := SetupTestDB(t)
		defer cleanup()
		store := NewDBStore(db)

		now := time.Now()
		db.Create(&AppSessionV1{ID: "s1", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s2", ApplicationID: "app1", SessionData: "{}", Status: app.AppSessionStatusClosed, Nonce: 2, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s3", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 1, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s4", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusOpen, Nonce: 2, UpdatedAt: now})
		db.Create(&AppSessionV1{ID: "s5", ApplicationID: "app2", SessionData: "{}", Status: app.AppSessionStatusClosed, Nonce: 3, UpdatedAt: now})

		results, err := store.CountActiveAppSessions(24 * time.Hour)
		require.NoError(t, err)

		countByLabel := make(map[string]uint64)
		for _, r := range results {
			countByLabel[r.Label] = r.Count
		}

		// CountActiveAppSessions counts all sessions regardless of status
		assert.Equal(t, uint64(2), countByLabel["app1"])
		assert.Equal(t, uint64(3), countByLabel["app2"])
	})
}
