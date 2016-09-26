package docker

import (
	"fmt"
	"math/rand"

	"github.com/docker/integreat/modules"
	"github.com/docker/integreat/types"
	"github.com/docker/integreat/util"

	"github.com/Sirupsen/logrus"
	"github.com/docker/docker/client"
)

func init() {
	modules.Register("docker", types.ModuleCreator(NewSuite))
}

type Suite struct {
	rand   *rand.Rand
	logger *logrus.Logger
	client *client.Client
}

func NewSuite(opts types.ModuleOpts) (types.Module, error) {
	docker, ok := opts.Config["docker"]
	if !ok {
		return nil, fmt.Errorf("docker config not found")
	}

	host, _ := docker["host"].(string)
	version, _ := docker["version"].(string)

	cli, err := client.NewClient(host, version, nil, map[string]string{})
	if err != nil {
		return nil, err
	}

	return &Suite{
		rand:   opts.Rand,
		logger: opts.Logger,
		client: cli,
	}, nil
}

func (s *Suite) GetCommand(cmd string) (types.TestCommand, error) {
	return modules.GetCommand(s, cmd)
}

func (s *Suite) CreateRandomImage(a types.TestArgs) (types.TestResult, error) {
	// TODO: shit.
}
