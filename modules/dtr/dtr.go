package dtr

import (
	"github.com/docker/integreat/modules"
	"github.com/docker/integreat/types"
)

func init() {
	modules.Register("dtr", types.ModuleCreator(NewSuite))
}

type Config struct {
	Host string
	User string
	Pass string
}

type Suite struct {
	config Config
}

func NewSuite(config map[string]interface{}) (types.Module, error) {
	return &Suite{
		config: Config{},
	}, nil
}

func (s *Suite) GetCommand(cmd string) (types.TestCommand, error) {
	return modules.GetCommand(s, cmd)
}

func (s *Suite) CreateUser(a types.TestArgs) (types.TestResult, error) {
	return nil, nil
}
