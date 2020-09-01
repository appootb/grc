package backend

import (
	"fmt"
)

const (
	ServicePrefix    = "service"
	ServiceOpsPrefix = "ops"
	ServiceNodeIDKey = "node_id"
)

func ServiceDiscoveryPrefixKey(path string) string {
	return fmt.Sprintf("%s/%s/", path, ServicePrefix)
}

func ServiceDiscoveryKey(path, service, node string) string {
	return fmt.Sprintf("%s/%s/%s/%s", path, ServicePrefix, service, node)
}

func ServiceOpsKey(path, service, node string) string {
	return fmt.Sprintf("%s/%s/%s/%s", path, ServiceOpsPrefix, service, node)
}

func ServiceNodeIDIncrKey(path, service string) string {
	return fmt.Sprintf("%s/%s/%s", path, ServiceNodeIDKey, service)
}
