package model

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/appootb/grc/backend"
	"github.com/appootb/grc/backend/etcd"
	"github.com/appootb/grc/dashboard/config"
)

var (
	provider backend.Provider
	services sync.Map
)

func Init() (err error) {
	switch config.GlobalConfig.Provider.Type {
	case backend.Etcd:
		provider, err = etcd.NewProvider(
			context.Background(),
			strings.Split(config.GlobalConfig.Provider.Endpoints, ","),
			config.GlobalConfig.Provider.Username,
			config.GlobalConfig.Provider.Password)
	default:
		err = errors.New("unknown backend type: " + config.GlobalConfig.Provider.Type)
	}
	if err != nil {
		return err
	}
	// Sync service list
	go syncServiceList()
	return
}

func syncServiceList() {
	m := NewService()
	ticker := time.NewTicker(time.Minute * 30)

	for {
		m.Sync()
		<-ticker.C
	}
}
