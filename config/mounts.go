package config

func (cfg *Config) DeployMounts() error {
	cfg.Debug("Enter pkg/config/mounts.go")

	req := cfg.NewRequest()
	req.Mounts = cfg.Mounts

	if err := req.Send(); err != nil {
		return err
	}

	cfg.Debug("Leave pkg/config/mounts.go")
	return nil
}
