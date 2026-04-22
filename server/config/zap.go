package config

type Zap struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Prefix string `mapstructure:"prefix"`
}
