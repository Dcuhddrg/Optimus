package initialize

import (
	"context"
	"time"

	systemService "optimus/server/service/system"
	persistence "optimus/server/service/system/store"

	"gorm.io/gorm"
)

func startAuditRetentionWorker(ctx context.Context, db *gorm.DB, retentionDays int) {
	if retentionDays <= 0 {
		retentionDays = 90
	}
	svc := systemService.NewAuditService(persistence.NewGormAuditStore(db))
	go runRetentionLoop(ctx, svc.PurgeBefore, retentionDays, nextDailyTick())
}

func runRetentionLoop(ctx context.Context, purge func(context.Context, time.Time) error, retentionDays int, firstTick time.Duration) {
	timer := time.NewTimer(firstTick)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			cutoff := time.Now().AddDate(0, 0, -retentionDays)
			_ = purge(ctx, cutoff)
			timer.Reset(nextDailyTick())
		}
	}
}

func nextDailyTick() time.Duration {
	now := time.Now()
	next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if !next.After(now) {
		next = next.Add(24 * time.Hour)
	}
	return time.Until(next)
}
