package initialize

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunRetentionLoopCallsPurgeAfterTick(t *testing.T) {
	purged := make(chan time.Time, 1)
	fakePurger := func(ctx context.Context, before time.Time) error {
		purged <- before
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runRetentionLoop(ctx, fakePurger, 30, 0)

	select {
	case before := <-purged:
		diff := time.Since(before)
		require.Greater(t, diff, 29*24*time.Hour)
		require.Less(t, diff, 31*24*time.Hour)
	case <-time.After(2 * time.Second):
		t.Fatal("purge was not called within timeout")
	}
}
