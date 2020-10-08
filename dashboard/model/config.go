package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/appootb/grc/backend"
	"github.com/appootb/grc/dashboard/config"
)

var (
	ServiceNotExist = errors.New("service not exist")
)

type Config struct {
	backend.ConfigItem

	Key string `json:"key,omitempty"`
}

func NewConfig() *Config {
	return &Config{}
}

func (m Config) GetKeys(service string) ([]*Config, error) {
	name, ok := services.Load(service)
	if !ok {
		return nil, ServiceNotExist
	}
	backendKey := backend.ServiceConfigKey(config.GlobalConfig.Provider.BasePath, name.(string))
	kvs, err := provider.Get(backendKey, true)
	if err != nil {
		return nil, err
	}
	items := make([]*Config, 0, len(kvs))
	for _, pair := range kvs {
		var item backend.ConfigItem
		if err := json.Unmarshal([]byte(pair.Value), &item); err != nil {
			return nil, err
		}
		key := strings.TrimPrefix(pair.Key, backendKey)
		items = append(items, &Config{
			ConfigItem: item,
			Key:        strings.ReplaceAll(key, "/", "."),
		})
	}
	return items, nil
}

func (m Config) UpdateKey(service, key string) error {
	m.Key = ""
	name, ok := services.Load(service)
	if !ok {
		return ServiceNotExist
	}
	backendKey := fmt.Sprintf("%s/%s/%s/%s",
		config.GlobalConfig.Provider.BasePath, backend.ConfigPrefix, name.(string), key)
	return provider.Set(backendKey, m.ConfigItem.String(), 0)
}

func (m Config) DeleteKey(service, key string) error {
	name, ok := services.Load(service)
	if !ok {
		return ServiceNotExist
	}
	backendKey := fmt.Sprintf("%s/%s/%s/%s",
		config.GlobalConfig.Provider.BasePath, backend.ConfigPrefix, name.(string), key)
	return provider.Delete(backendKey, false)
}
