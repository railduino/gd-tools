package main

import (
	"gopkg.in/yaml.v3"
)

// TODO es fehlt noch die erzeugung von networks !!!

// within services
type NetworkNames []string

// at compose top-level
type NetworkConfig struct {
	External bool   `yaml:"external,omitempty"`
	Driver   string `yaml:"driver,omitempty"`
	Name     string
}

type NetworkList map[string]NetworkConfig

func (n *NetworkNames) UnmarshalYAML(value *yaml.Node) error {
	var list []string
	if err := value.Decode(&list); err != nil {
		return err
	}

	*n = list
	return nil
}

func (nl *NetworkList) UnmarshalYAML(value *yaml.Node) error {
	raw := make(map[string]NetworkConfig)
	if err := value.Decode(&raw); err != nil {
		return err
	}

	for name, net := range raw {
		net.Name = name
		raw[name] = net
	}

	*nl = raw
	return nil
}
