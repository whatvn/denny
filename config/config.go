package config

import (
	goconfig "github.com/whatvn/denny/go_config"
	"github.com/whatvn/denny/go_config/source"
	"github.com/whatvn/denny/go_config/source/env"
	"github.com/whatvn/denny/go_config/source/file"
)

var (
	cfg Config
)

type Config goconfig.Config

// New will load config from various config file file in Json format
// if same config param available in environment, environment param will
// take higher priority
func New(sources ...string) error {
	cfg = goconfig.NewConfig()
	if len(sources) == 0 {
		return cfg.Load(env.NewSource())
	}
	var cfgSources []source.Source
	for _, s := range sources {
		cfgSources = append(cfgSources, file.NewSource(file.WithPath(s)))
	}
	if err := cfg.Load(cfgSources...); err != nil {
		return err
	}

	return cfg.Load(env.NewSource())
}

func Reload() error {
	return cfg.Sync()
}

func GetString(path ...string) string {
	return cfg.Get(path...).String("")
}

func GetStringMap(path ...string) map[string]string {
	return cfg.Get(path...).StringMap(nil)
}

func GetStringArray(path ...string) []string {
	return cfg.Get(path...).StringSlice(nil)
}

func GetInt(path ...string) int {
	return cfg.Get(path...).Int(0)
}

func Scan(t interface{}, path ...string) error {
	return cfg.Get(path...).Scan(t)
}

func Map() map[string]interface{} {
	return cfg.Map()
}
