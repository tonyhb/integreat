package config

import (
	"github.com/docker/integreat/types"

	"gopkg.in/yaml.v2"
)

type Configuration struct {
	Base    Base
	Modules []string
	Config  map[string]map[string]interface{}
	Tests   []Test
}

type Base struct {
	Version int
}

type Test struct {
	Id string

	// Name represents the human-readable name of the test
	Name string

	// Command describes the suite and function to call, in the format of
	// `Suite::FunctionName`
	Command string

	// Args is a map of arguments passed to the test command
	Args types.TestArgs

	// Repeat represents how many times this test will be repeated in sequence.
	// The default is 1.
	Repeat int
}

func Parse(data []byte) (*Configuration, error) {
	c := new(Configuration)
	err := yaml.Unmarshal(data, c)
	return c, err
}
