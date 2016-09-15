package dtr

import (
	"github.com/docker/integreat/modules"
)

func init() {
	modules.Register("dtr", NewSuite)
}

type Config struct {
	Host string
	User string
	Pass string
}

type Suite struct {
	config Config
}

func NewSuite(config map[string]interface{}) (*Suite, error) {
	return &Suite{
		config: Config{},
	}, nil
}

func (s *Suite) GetCommand(cmd string) (TestCommand, error) {
	return modules.Get(s, cmd)
}

func (s *Suite) CreateUser(a TestArgs) (TestResult, error) {
	return nil, nil
}
