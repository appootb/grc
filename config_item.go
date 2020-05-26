package grc

import "encoding/json"

const (
	ConfigPrefix = "config"
)

type ConfigItem struct {
	Type    string `json:"type"`
	Value   string `json:"value"`
	Comment string `json:"comment"`
}

type ConfigItems map[string]ConfigItem

func (c ConfigItems) Add(items ConfigItems) {
	for k, v := range items {
		c[k] = v
	}
}

func (c ConfigItems) BackendKVs(servicePath string) map[string]string {
	kvs := make(map[string]string, len(c))
	for key, item := range c {
		v, _ := json.Marshal(item)
		kvs[servicePath+key] = string(v)
	}
	return kvs
}
