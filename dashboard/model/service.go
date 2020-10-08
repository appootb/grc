package model

import (
	"fmt"
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

func (m Service) GetNames() []string {
	var names []string
	services.Range(func(name, _ interface{}) bool {
		names = append(names, name.(string))
		return true
	})
	return names
}
