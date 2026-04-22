package initialize

import (
	"errors"
	"testing"

	"optimus/server/config"
	"optimus/server/global"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"

	"github.com/stretchr/testify/require"
)

func TestBuildMySQLDSN(t *testing.T) {
	cfg := config.Mysql{
		Username: "root",
		Password: "secret",
		Path:     "127.0.0.1",
		Port:     3306,
		DBName:   "optimus",
		Config:   "charset=utf8mb4&parseTime=True&loc=Local",
	}

	require.Equal(t, "root:secret@tcp(127.0.0.1:3306)/optimus?charset=utf8mb4&parseTime=True&loc=Local", buildMySQLDSN(cfg))
}

func TestConfigureMySQLPoolClosesOnPingFailure(t *testing.T) {
	conn := &stubSQLDB{pingErr: errors.New("ping failed")}

	err := configureMySQLPool(conn, config.Mysql{MaxIdleConns: 10, MaxOpenConns: 20})

	require.Error(t, err)
	require.True(t, conn.closed)
	require.Equal(t, 10, conn.maxIdle)
	require.Equal(t, 20, conn.maxOpen)
}

func TestGormLogsUnsupportedDBType(t *testing.T) {
	core, logs := observer.New(zap.ErrorLevel)
	logger := zap.New(core)

	previousConfig := global.OPTIMUS_CONFIG
	previousLogger := global.OPTIMUS_LOG
	previousDB := global.OPTIMUS_DB
	t.Cleanup(func() {
		global.OPTIMUS_CONFIG = previousConfig
		global.OPTIMUS_LOG = previousLogger
		global.OPTIMUS_DB = previousDB
	})

	global.OPTIMUS_CONFIG = config.Config{
		System: config.System{DBType: "postgres"},
	}
	global.OPTIMUS_LOG = logger
	global.OPTIMUS_DB = nil

	db := Gorm()

	require.Nil(t, db)
	require.Nil(t, global.OPTIMUS_DB)
	entries := logs.All()
	require.Len(t, entries, 1)
	require.Equal(t, "unsupported db type", entries[0].Message)
	require.Equal(t, "postgres", entries[0].ContextMap()["db_type"])
}

type stubSQLDB struct {
	maxIdle int
	maxOpen int
	closed  bool
	pingErr error
}

func (s *stubSQLDB) SetMaxIdleConns(n int) {
	s.maxIdle = n
}

func (s *stubSQLDB) SetMaxOpenConns(n int) {
	s.maxOpen = n
}

func (s *stubSQLDB) Ping() error {
	return s.pingErr
}

func (s *stubSQLDB) Close() error {
	s.closed = true
	return nil
}
