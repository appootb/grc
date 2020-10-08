package config

import (
	"flag"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

var (
	FilePath     string
	GlobalConfig Config
)

type Base struct {
	ServeAddress  string `yaml:"address"`
	AdminPassword string `yaml:"admin-pass"`
}

type LDAP struct {
	Enable  bool   `yaml:"enable"`
	Address string `yaml:"address"`
	BaseDN  string `yaml:"base-dn"`
}

type Backend struct {
	Type      string `yaml:"type"`
	BasePath  string `yaml:"base-path"`
	Endpoints string `yaml:"endpoints"`
	Username  string `yaml:"username"`
	Password  string `yaml:"password"`
}

type Config struct {
	Dashboard Base    `yaml:"dashboard"`
	LDAP      LDAP    `yaml:"ldap"`
	Provider  Backend `yaml:"provider"`
}

func Init() error {
	flag.StringVar(&FilePath, "cfg", "./config.yaml", "yaml config file path")
	flag.Parse()

	cfg, err := ioutil.ReadFile(FilePath)
	if err != nil {
		panic("read config file failed, err: " + err.Error())
	}
	if err = yaml.Unmarshal(cfg, &GlobalConfig); err != nil {
		panic("parse config file failed, err: " + err.Error())
	}
	return nil
}
