package model

import (
	"fmt"
	"sort"
	"strings"

	"github.com/appootb/grc/backend"
	"github.com/appootb/grc/dashboard/config"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (m Service) Sync() {
	prefix := fmt.Sprintf("%s/%s/", config.GlobalConfig.Provider.BasePath, backend.ConfigPrefix)
	kvs, err := provider.Get(prefix, true)
	if err != nil {
		return
	}
	for _, kv := range kvs {
		key := strings.TrimPrefix(kv.Key, prefix)
		service := strings.Split(key, "/")[0]
		if service == "" {
			continue
		}
		name := strings.Title(strings.ToLower(strings.TrimPrefix(service, "COMPONENT_")))
		services.Store(name, service)
	}
}

func (m Service) Delete(service string) error {
	name, ok := services.Load(service)
	if !ok {
		return ServiceNotExist
	}
	services.Delete(service)
	// Config key prefix
	backendKey := fmt.Sprintf("%s/%s/%s/",
		config.GlobalConfig.Provider.BasePath, backend.ConfigPrefix, name.(string))
	if err := provider.Delete(backendKey, true); err != nil {
		return err
	}
	// Node ID key
	nodeIDKey := fmt.Sprintf("%s/%s/%s",
		config.GlobalConfig.Provider.BasePath, backend.ServiceNodeIDKey, name.(string))
	return provider.Delete(nodeIDKey, false)
}

func (m Service) GetNames() []string {
	var names []string
	services.Range(func(name, _ interface{}) bool {
		names = append(names, name.(string))
		return true
	})
	sort.Strings(names)
	return names
}
