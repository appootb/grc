# grc

grc (golang remote configuration) is a distribute service config and discovery center.

<p>
    <a href="https://hub.docker.com/r/appootb/grc-dashboard" alt="Activity">
        <img src="https://img.shields.io/docker/cloud/automated/appootb/grc-dashboard.svg" /></a>
    <a href="https://hub.docker.com/r/appootb/grc-dashboard" alt="Activity">
        <img src="https://img.shields.io/docker/cloud/build/appootb/grc-dashboard.svg" /></a>
    <img src="https://img.shields.io/github/license/appootb/grc.svg" />
</p>

## Features

### Backend supported

* [x] Memory - debug only
* [x] ETCD
* [ ] Consul
* [ ] ZooKeeper
* [ ] Redis

## Dashboard

### Install

* Update config file `dashboard/config.yaml`

* Run dashboard

`docker run -d -p 8000:80 -v :/etc/config.yaml appootb/grc-dashboard`

* [Optional] run dashboard and etcd with docker-compose

> docker-compose.yml

```yaml
version: '3'
services:
  etcd:
    image: quay.io/coreos/etcd:v3.3.25
    environment:
      - ETCDCTL_API=3
    command:
      - "/usr/local/bin/etcd"
      - "--name"
      - "s1"
      - "--data-dir"
      - "/etcd-data"
      - "--listen-client-urls"
      - "http://0.0.0.0:2379"
      - "--advertise-client-urls"
      - "http://0.0.0.0:2379"
      - "--listen-peer-urls"
      - "http://0.0.0.0:2380"
      - "--initial-advertise-peer-urls"
      - "http://0.0.0.0:2380"
      - "--initial-cluster"
      - "s1=http://0.0.0.0:2380"
      - "--initial-cluster-token"
      - "tkn"
      - "--initial-cluster-state"
      - "new"
    ports:
      - "2379:2379"

  dashboard:
    image: appootb/grc-dashboard
    ports:
      - "8000:80"
    volumes:
      - "./dashboard/config.yaml:/etc/config.yaml"
```

* Browser [http://localhost:8000](http://localhost:8000)

## Example

1. Initialize grc instance

```go
grc, _ = grc.New(WithDebugProvider(),
	WithConfigAutoCreation(),
	WithBasePath("/test"))
```

2. Define configuration structure

```go
type Config struct {
	//
	// Static configs, won't been changed when process running
	// Internal type
	IV  int                         `comment:"comment for dashboard" default:"1"`
	// Pointer
	PIV *int                        `comment:"comment for dashboard" default:"10"`
	// Map
	MV  map[string]int              `comment:"comment for dashboard" default:"a:1,b:2"`
	EMV map[string]map[string]int32 `comment:"comment for dashboard" default:"a_1:bb_2:1,cc_2:2;b_1:dd_2:19,ee_2:20"`
	// Slice
	AV  []bool                      `comment:"comment for dashboard" default:"false,true"`
	EAV [][]string                  `comment:"comment for dashboard" default:"a_1,b_1,c_1;a_2,b_2,c_2"`

	//
	// Dynamic configs, should be updated automatically
	DIV grc.Int   `comment:"comment for dashboard" default:"2"`
	DMV grc.Map   `comment:"comment for dashboard" default:"a_1:bb_2:1,cc_2:2;b_1:dd_2:19,ee_2:20"`
	DAV grc.Array `comment:"comment for dashboard" default:"a_1,b_1,c_1;a_2,b_2,c_2"`
}
``` 

3. Register service configuration

```go
var cfg Config
err := grc.RegisterConfig("Service_Name", &cfg)
if err != nil {
	// TODO
}
```

4. That's all, use `cfg` directly which is concurrency safe.