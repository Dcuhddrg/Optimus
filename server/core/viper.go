package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"optimus/server/config"
	"optimus/server/global"

	"github.com/spf13/viper"
)

const envPrefix = "OPTIMUS"

func Viper() error {
	configPath, err := resolveDefaultConfigPath()
	if err != nil {
		return err
	}

	cfg, vp, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	global.OPTIMUS_CONFIG = cfg
	global.OPTIMUS_VP = vp
	return nil
}

func LoadConfigForTest(configPath string) (config.Config, error) {
	cfg, vp, err := loadConfig(configPath)
	if err != nil {
		return config.Config{}, err
	}

	global.OPTIMUS_CONFIG = cfg
	global.OPTIMUS_VP = vp
	return cfg, nil
}

func loadConfig(configPath string) (config.Config, *viper.Viper, error) {
	vp := viper.New()
	vp.SetConfigFile(filepath.Clean(configPath))
	vp.SetEnvPrefix(envPrefix)
	vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	vp.AutomaticEnv()

	if err := vp.ReadInConfig(); err != nil {
		return config.Config{}, nil, err
	}

	cfg := config.Config{}
	if err := vp.Unmarshal(&cfg); err != nil {
		return config.Config{}, nil, err
	}

	return cfg, vp, nil
}

func resolveDefaultConfigPath() (string, error) {
	for _, candidate := range []string{
		"config.yaml",
		filepath.Join("server", "config.yaml"),
	} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("config.yaml not found in current directory or server/config.yaml")
}
