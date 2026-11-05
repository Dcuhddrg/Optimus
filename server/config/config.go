package config

type Config struct {
	System System `mapstructure:"system"`
	Cloud  Cloud  `mapstructure:"cloud"`
	Zap    Zap    `mapstructure:"zap"`
	Mysql  Mysql  `mapstructure:"mysql"`
	JWT    JWT    `mapstructure:"jwt"`
}
