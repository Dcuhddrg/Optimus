package config

type System struct {
	Env          string `mapstructure:"env"`
	Addr         int    `mapstructure:"addr"`
	DBType       string `mapstructure:"db-type"`
	UseRedis     bool   `mapstructure:"use-redis"`
	UseCasbin    bool   `mapstructure:"use-casbin"`
	RouterPrefix string `mapstructure:"router-prefix"`
}
