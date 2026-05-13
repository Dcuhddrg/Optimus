package initialize

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunRetentionLoopCallsPurgeAfterTick(t *testing.T) {
	var purgeCalled bool
	var purgedBefore time.Time

	fakePurger := func(ctx context.Context, before time.Time) error {
		purgeCalled = true
		purgedBefore = before
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runRetentionLoop(ctx, fakePurger, 30, 0)

	time.Sleep(50 * time.Millisecond)
	cancel()

	require.True(t, purgeCalled)
	diff := time.Since(purgedBefore)
	require.Greater(t, diff, 29*24*time.Hour)
	require.Less(t, diff, 31*24*time.Hour)
}
