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

func RunServer() {
	addr := fmt.Sprintf(":%d", global.OPTIMUS_CONFIG.System.Addr)
	server := &http.Server{
		Addr:              addr,
		Handler:           http.NewServeMux(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	_, _ = os.Stdout.WriteString(Banner(global.Version, global.OPTIMUS_CONFIG.System.Addr))
	Logger().Info("starting http server", zap.String("addr", addr), zap.String("version", global.Version))

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		Logger().Error("http server stopped", zap.Error(err))
		panic(err)
	}
}
