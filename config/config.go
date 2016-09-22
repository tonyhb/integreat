package config

import (
	"github.com/docker/integreat/types"

	"gopkg.in/yaml.v2"
)

func Parse(data []byte) (*types.Configuration, error) {
	c := new(types.Configuration)
	err := yaml.Unmarshal(data, c)
	return c, err
}
