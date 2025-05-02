package main

import (
	"fmt"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// KVList ist ein flexibler Typ für Environment, Labels, Ports, Volumes usw.
type KVList map[string]string

// UnmarshalYAML unterstützt sowohl MappingNode als auch SequenceNode ("key=value" oder "key:value")
func (l *KVList) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.MappingNode:
		var m map[string]string
		if err := value.Decode(&m); err != nil {
			return err
		}
		*l = m
		return nil

	case yaml.SequenceNode:
		m := make(map[string]string)
		for _, item := range value.Content {
			var s string
			if err := item.Decode(&s); err != nil {
				return err
			}
			// intelligent splitten: erst versuchen "=" dann ":"
			if parts := strings.SplitN(s, "=", 2); len(parts) == 2 {
				m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			} else if parts := strings.SplitN(s, ":", 2); len(parts) == 2 {
				m[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			} else {
				msg := T("yaml-err-invalid-kvlist")
				return fmt.Errorf(msg)
			}
		}
		*l = m
		return nil

	default:
		msg := T("yaml-err-unexpected-kvlist")
		return fmt.Errorf(msg)
	}
}

// MarshalYAML schreibt die KVList immer als sortierte Liste ("key=value")
func (l KVList) MarshalYAML() (interface{}, error) {
	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	list := make([]string, 0, len(l))
	for _, k := range keys {
		list = append(list, fmt.Sprintf("%s=%s", k, l[k]))
	}
	return list, nil
}

// Get holt den Wert zu einem Key. Gibt "" zurück, falls Key nicht existiert.
func (l KVList) Get(key string) string {
	return l[key]
}

// Set setzt oder überschreibt einen Key mit einem Wert.
func (l KVList) Set(key, value string) {
	l[key] = value
}

// Delete entfernt einen Key, falls vorhanden.
func (l KVList) Delete(key string) {
	delete(l, key)
}

// Has prüft, ob ein Key existiert.
func (l KVList) Has(key string) bool {
	_, ok := l[key]
	return ok
}

// Keys gibt eine sortierte Liste aller Keys zurück.
func (l KVList) Keys() []string {
	keys := make([]string, 0, len(l))
	for k := range l {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
