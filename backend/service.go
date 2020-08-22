package backend

import (
	"fmt"
)

const (
	ServicePrefix       = "service"
	TrafficWeightPrefix = "weight"
)

func ServiceDiscoveryPrefixKey(path string) string {
	return fmt.Sprintf("%s/%s/", path, ServicePrefix)
}

func ServiceDiscoveryKey(path, service, node string) string {
	return fmt.Sprintf("%s/%s/%s/%s", path, ServicePrefix, service, node)
}

func TrafficWeightKey(path, service, node string) string {
	return fmt.Sprintf("%s/%s/%s/%s", path, TrafficWeightPrefix, service, node)
}
