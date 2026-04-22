package initialize

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"optimus/server/global"

	"github.com/casbin/casbin/v2"
	"github.com/casbin/casbin/v2/model"
)

func Casbin() *casbin.Enforcer {
	global.OPTIMUS_ENFORCER = nil

	if !casbinEnabled() {
		return nil
	}

	modelPath, err := resolveCasbinModelPath(global.OPTIMUS_CONFIG.Casbin.ModelPath)
	if err != nil {
		logDependencyInitFailure("casbin", "resolve model path", err)
		return nil
	}

	casbinModel, err := model.NewModelFromFile(modelPath)
	if err != nil {
		logDependencyInitFailure("casbin", "load model file", err)
		return nil
	}

	enforcer, err := casbin.NewEnforcer(casbinModel)
	if err != nil {
		logDependencyInitFailure("casbin", "create enforcer", err)
		return nil
	}

	global.OPTIMUS_ENFORCER = enforcer
	return enforcer
}

func casbinEnabled() bool {
	return global.OPTIMUS_CONFIG.System.UseCasbin && global.OPTIMUS_CONFIG.Casbin.Enable
}

func resolveCasbinModelPath(configuredPath string) (string, error) {
	trimmedPath := strings.TrimSpace(configuredPath)
	if trimmedPath == "" {
		return "", fmt.Errorf("casbin model path is empty")
	}

	cleanPath := filepath.Clean(trimmedPath)
	for _, candidate := range casbinModelPathCandidates(cleanPath) {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return cleanPath, nil
}

func casbinModelPathCandidates(cleanPath string) []string {
	candidates := []string{cleanPath}
	if !strings.HasPrefix(cleanPath, "server"+string(filepath.Separator)) {
		candidates = append(candidates, filepath.Join("server", cleanPath))
	}
	return candidates
}
