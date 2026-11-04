package cloud

// ServiceGroup embeds the stateless cloud account and asset services.
type ServiceGroup struct {
	CloudAccountService
	AssetRegistryService
}
