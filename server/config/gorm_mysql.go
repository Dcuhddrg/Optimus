package config

type Mysql struct {
	Path         string `mapstructure:"path"`
	Port         int    `mapstructure:"port"`
	DBName       string `mapstructure:"db-name"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	Config       string `mapstructure:"config"`
	MaxIdleConns int    `mapstructure:"max-idle-conns"`
	MaxOpenConns int    `mapstructure:"max-open-conns"`
	LogMode      string `mapstructure:"log-mode"`
}
