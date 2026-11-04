package config

// JWT mirrors GVA: lifetimes are duration strings such as "7d" or "1h",
// parsed with utils.ParseDuration.
type JWT struct {
	SigningKey  string `mapstructure:"signing-key"`
	ExpiresTime string `mapstructure:"expires-time"`
	BufferTime  string `mapstructure:"buffer-time"`
	Issuer      string `mapstructure:"issuer"`
}
