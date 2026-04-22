package core

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"optimus/server/global"
)

func TestViperLoadsConfigFileAndEnvOverride(t *testing.T) {
	global.OPTIMUS_VP = nil
	t.Setenv("OPTIMUS_SYSTEM_ADDR", "19091")

	if global.OPTIMUS_VP != nil {
		t.Fatal("expected OPTIMUS_VP to be nil before bootstrap")
	}

	cfg, err := LoadConfigForTest("../config.yaml")
	if err != nil {
		t.Fatalf("LoadConfigForTest returned error: %v", err)
	}

	if cfg.System.Addr != 19091 {
		t.Fatalf("expected System.Addr to be 19091, got %d", cfg.System.Addr)
	}

	if cfg.System.DBType != "mysql" {
		t.Fatalf("expected System.DBType to be mysql, got %q", cfg.System.DBType)
	}

	if global.OPTIMUS_VP == nil {
		t.Fatal("expected OPTIMUS_VP to be initialized")
	}

	if got := global.OPTIMUS_VP.GetString("system.db-type"); got != "mysql" {
		t.Fatalf("expected Viper to retain system.db-type mysql, got %q", got)
	}
}

func TestViperLoadsDefaultConfigFromServerDirAndRepoRoot(t *testing.T) {
	serverDir, repoRoot := testProjectPaths(t)

	for _, workingDir := range []string{serverDir, repoRoot} {
		t.Run(workingDir, func(t *testing.T) {
			global.OPTIMUS_VP = nil

			previousWD, err := os.Getwd()
			if err != nil {
				t.Fatalf("Getwd returned error: %v", err)
			}
			t.Cleanup(func() {
				_ = os.Chdir(previousWD)
			})

			if err := os.Chdir(workingDir); err != nil {
				t.Fatalf("Chdir(%q) returned error: %v", workingDir, err)
			}

			if err := Viper(); err != nil {
				t.Fatalf("Viper returned error from %q: %v", workingDir, err)
			}

			if global.OPTIMUS_VP == nil {
				t.Fatalf("expected OPTIMUS_VP to be initialized from %q", workingDir)
			}

			if got := global.OPTIMUS_CONFIG.System.DBType; got != "mysql" {
				t.Fatalf("expected System.DBType mysql from %q, got %q", workingDir, got)
			}
		})
	}
}

func testProjectPaths(t *testing.T) (string, string) {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}

	coreDir := filepath.Dir(file)
	serverDir := filepath.Dir(coreDir)
	repoRoot := filepath.Dir(serverDir)
	return serverDir, repoRoot
}
