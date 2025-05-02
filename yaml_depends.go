package main

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type DependsOn []string

func (d *DependsOn) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		var list []string
		if err := value.Decode(&list); err != nil {
			return err
		}
		*d = list
		return nil

	case yaml.MappingNode:
		var raw map[string]interface{}
		if err := value.Decode(&raw); err != nil {
			return err
		}
		var list []string
		for k := range raw {
			list = append(list, k)
		}
		*d = list
		return nil

	default:
		msg := T("yaml-err-invalid-depends")
		return fmt.Errorf(msg)
	}
}
