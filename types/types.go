package types

import (
	"math/rand"

	"github.com/Sirupsen/logrus"
)

// TestCommand is the function signature of valid test commands callable from
// within a YAML file
type TestCommand func(TestArgs) (TestResult, error)

type TestResult map[string]interface{}

// Module is an interface representing a registerable suite of test commands
// for a specific product.
type Module interface {
	GetCommand(string) (TestCommand, error)
}

type ModuleOpts struct {
	Config ModuleConfig

	Logger *logrus.Logger
	Rand   *rand.Rand
}

// ModuleCreator is a function which returns a concrete Module or an error
// given a map of config options.
//
// These config options should be returned from the specific config key of
// the main YAML config file.
type ModuleCreator func(ModuleOpts) (Module, error)

// TestArgs is a map of interfaces representing arguments for a test command
// defined in the YAML file
type TestArgs map[string]interface{}

func (t TestArgs) Bool(key string) bool {
	b, ok := t[key].(bool)
	if !ok {
		return false
	}
	return b
}

func (t TestArgs) String(key string) string {
	s, ok := t[key].(string)
	if !ok {
		return ""
	}
	return s
}
