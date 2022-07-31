package grc

import (
	"context"

	"github.com/appootb/grc/backend"
	"github.com/appootb/grc/backend/etcd"
	"github.com/appootb/grc/backend/memory"
)

// Option interface sets options such as provider, autoCreation, etc.
type Option interface {
	apply(*RemoteConfig)
}

// funcServerOption wraps a function that modifies serverOptions into an
// implementation of the ServerOption interface.
type funcServerOption struct {
	f func(*RemoteConfig)
}

func (fdo *funcServerOption) apply(do *RemoteConfig) {
	fdo.f(do)
}

func newFuncServerOption(f func(*RemoteConfig)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

func WithContext(ctx context.Context) Option {
	return newFuncServerOption(func(rc *RemoteConfig) {
		rc.ctx = ctx
	})
}

func WithConfigAutoCreation() Option {
	return newFuncServerOption(func(rc *RemoteConfig) {
		rc.autoCreation = true
	})
}

func WithBasePath(path string) Option {
	return newFuncServerOption(func(rc *RemoteConfig) {
		rc.path = path
	})
}

func WithProvider(provider backend.Provider) Option {
	return newFuncServerOption(func(rc *RemoteConfig) {
		rc.provider = provider
	})
}

func WithDebugProvider() Option {
	return newFuncServerOption(func(rc *RemoteConfig) {
		rc.provider = memory.NewProvider()
	})
}

func WithEtcdProvider(ctx context.Context, endPoints []string, username, password string) Option {
	return newFuncServerOption(func(rc *RemoteConfig) {
		provider, err := etcd.NewProvider(ctx, endPoints, username, password)
		if err != nil {
			panic("grc: connect to etcd failed: " + err.Error())
		}
		rc.provider = provider
	})
}

func WithCallbackManger(mgr Callback) Option {
	return newFuncServerOption(func(_ *RemoteConfig) {
		callbackMgr = mgr
	})
}
