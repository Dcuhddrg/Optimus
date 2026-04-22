package config

type JWT struct {
	SigningKey  string `mapstructure:"signing-key"`
	ExpiresTime int    `mapstructure:"expires-time"`
	BufferTime  int    `mapstructure:"buffer-time"`
	Issuer      string `mapstructure:"issuer"`
}
