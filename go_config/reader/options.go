package reader

import (
	"github.com/whatvn/denny/go_config/encoder"
	"github.com/whatvn/denny/go_config/encoder/hcl"
	"github.com/whatvn/denny/go_config/encoder/json"
	"github.com/whatvn/denny/go_config/encoder/toml"
	"github.com/whatvn/denny/go_config/encoder/xml"
	"github.com/whatvn/denny/go_config/encoder/yaml"
)

type Options struct {
	Encoding map[string]encoder.Encoder
}

type Option func(o *Options)

func NewOptions(opts ...Option) Options {
	options := Options{
		Encoding: map[string]encoder.Encoder{
			"json": json.NewEncoder(),
			"yaml": yaml.NewEncoder(),
			"toml": toml.NewEncoder(),
			"xml":  xml.NewEncoder(),
			"hcl":  hcl.NewEncoder(),
			"yml":  yaml.NewEncoder(),
		},
	}
	for _, o := range opts {
		o(&options)
	}
	return options
}

func WithEncoder(e encoder.Encoder) Option {
	return func(o *Options) {
		if o.Encoding == nil {
			o.Encoding = make(map[string]encoder.Encoder)
		}
		o.Encoding[e.String()] = e
	}
}
