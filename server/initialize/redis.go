package initialize

import (
	"context"
	"time"

	"optimus/server/global"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const redisPingTimeout = 2 * time.Second

func Redis() *redis.Client {
	global.OPTIMUS_REDIS = nil

	if !global.OPTIMUS_CONFIG.System.UseRedis {
		return nil
	}

	client := redis.NewClient(&redis.Options{
		Addr:     global.OPTIMUS_CONFIG.Redis.Addr,
		Password: global.OPTIMUS_CONFIG.Redis.Password,
		DB:       global.OPTIMUS_CONFIG.Redis.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), redisPingTimeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logDependencyInitFailure("redis", "ping redis", err)
		_ = client.Close()
		return nil
	}

	global.OPTIMUS_REDIS = client
	return client
}

func logDependencyInitFailure(component, action string, err error) {
	logger, ok := global.OPTIMUS_LOG.(*zap.Logger)
	if !ok || logger == nil {
		return
	}

	logger.Error(component+" init failed", zap.String("action", action), zap.Error(err))
}
