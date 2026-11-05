package config

type System struct {
	Env                   string `mapstructure:"env"`
	Addr                  int    `mapstructure:"addr"`
	DBType                string `mapstructure:"db-type"`
	RouterPrefix          string `mapstructure:"router-prefix"`
	AdminInitialPassword  string `mapstructure:"admin-initial-password"`
	AuditLogRetentionDays int    `mapstructure:"audit-log-retention-days"`
}
