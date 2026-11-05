package core

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"optimus/server/config"
	"optimus/server/global"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const envPrefix = "OPTIMUS"

func Viper() *viper.Viper {
	cfg, vp, err := loadConfig(getConfigPath())
	if err != nil {
		panic(err)
	}

	global.OPTIMUS_CONFIG = cfg
	return vp
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

const (
	// ConfigEnv names the environment variable selecting the config file.
	ConfigEnv         = "OPTIMUS_CONFIG"
	ConfigDefaultFile = "config.yaml"
	ConfigTestFile    = "config.test.yaml"
	ConfigDebugFile   = "config.debug.yaml"
	ConfigReleaseFile = "config.release.yaml"
)

// getConfigPath selects the config file as GVA: the -c flag, then the
// OPTIMUS_CONFIG environment variable, then a file for the current gin mode,
// then config.yaml. Optimus also accepts server/config.yaml when started from
// the repository root.
func getConfigPath() string {
	if flag.Lookup("c") == nil {
		flag.String("c", "", "choose config file.")
	}
	if !flag.Parsed() {
		flag.Parse()
	}
	if config := flag.Lookup("c").Value.String(); config != "" {
		fmt.Printf("using config file from the -c flag: %s\n", config)
		return config
	}
	if config := os.Getenv(ConfigEnv); config != "" {
		fmt.Printf("using config file from %s: %s\n", ConfigEnv, config)
		return config
	}

	var config string
	switch gin.Mode() {
	case gin.DebugMode:
		config = ConfigDebugFile
	case gin.ReleaseMode:
		config = ConfigReleaseFile
	case gin.TestMode:
		config = ConfigTestFile
	}
	if _, err := os.Stat(config); err == nil {
		fmt.Printf("using config file for gin %s mode: %s\n", gin.Mode(), config)
		return config
	}
	for _, candidate := range []string{ConfigDefaultFile, filepath.Join("server", ConfigDefaultFile)} {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ConfigDefaultFile
}
