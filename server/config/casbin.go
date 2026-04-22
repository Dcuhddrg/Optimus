package config

type Casbin struct {
	Enable       bool   `mapstructure:"enable"`
	ModelPath    string `mapstructure:"model-path"`
	PolicySource string `mapstructure:"policy-source"`
}
