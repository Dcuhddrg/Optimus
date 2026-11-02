package cloud

// ServiceGroup holds the initialized cloud account and asset services.
type ServiceGroup struct {
	*CloudAccountService
	*AssetRegistryService
}
