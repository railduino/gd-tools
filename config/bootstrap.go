package config

func (cfg *Config) DeployBootstrap() error {
	cfg.Debug("Enter pkg/config/bootstrap.go")

	req := cfg.NewRequest()
	req.FQDN = cfg.FQDN()
	req.TimeZone = cfg.TimeZone
	req.SwapSize = cfg.SwapSize

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/bootstrap.go")
	return nil
}
