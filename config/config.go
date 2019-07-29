package config

import (
	goconfig "github.com/micro/go-config"
	"github.com/micro/go-config/source"
	"github.com/micro/go-config/source/env"
	"github.com/micro/go-config/source/file"
)

var (
	cfg Config
)

type Config goconfig.Config

func New(sources ...string) error {
	cfg = goconfig.NewConfig()
	if len(sources) == 0 {
		return cfg.Load(env.NewSource())
	}
	cfgSources := []source.Source{}
	for _, s := range sources {
		cfgSources = append(cfgSources, file.NewSource(file.WithPath(s)))
	}
	if err := cfg.Load(cfgSources...); err != nil {
		return err
	}
	return cfg.Load(env.NewSource())
}

func GetString(path ...string) string {
	return cfg.Get(path...).String("")
}

func GetInt(path ...string) int {
	return cfg.Get(path...).Int(0)
}

func Scan(t interface{}, path ...string) error {
	return cfg.Get(path...).Scan(t)
}
