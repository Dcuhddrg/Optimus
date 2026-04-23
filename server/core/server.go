package core

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"optimus/server/global"

	"go.uber.org/zap"
)

func RunServer(handler http.Handler) {
	addr := fmt.Sprintf(":%d", global.OPTIMUS_CONFIG.System.Addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	_, _ = os.Stdout.WriteString(Banner(global.Version, global.OPTIMUS_CONFIG.System.Addr))
	Logger().Info("starting http server", zap.String("addr", addr), zap.String("version", global.Version))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		Logger().Error("http server stopped", zap.Error(err))
		panic(err)
	}
}
