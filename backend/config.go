package backend

import (
	"encoding/json"
	"fmt"
)

const (
	ConfigPrefix = "config"
)

func ServiceConfigKey(path, service string) string {
	return fmt.Sprintf("%s/%s/%s/", path, ConfigPrefix, service)
}

type ConfigItem struct {
	Type     string `json:"type"`
	HintType string `json:"hint_type"`
	Value    string `json:"value"`
	Comment  string `json:"comment"`
}

func (c ConfigItem) String() string {
	v, _ := json.Marshal(c)
	return string(v)
}

type ConfigItems map[string]*ConfigItem

func (c ConfigItems) Add(items ConfigItems) {
	for k, v := range items {
		c[k] = v
	}
}

func (c ConfigItems) KVs(servicePath string) map[string]string {
	kvs := make(map[string]string, len(c))
	for key, item := range c {
		kvs[servicePath+key] = item.String()
	}
	return kvs
}
