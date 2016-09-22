package dtr

import (
	"fmt"
	"math/rand"

	"github.com/docker/integreat/modules"
	"github.com/docker/integreat/modules/dtr/client"
	"github.com/docker/integreat/types"
	"github.com/docker/integreat/util"

	"github.com/Sirupsen/logrus"
)

func init() {
	modules.Register("dtr", types.ModuleCreator(NewSuite))
}

type Suite struct {
	rand   *rand.Rand
	logger *logrus.Logger
	client client.Client
}

func NewSuite(opts types.ModuleOpts) (types.Module, error) {
	dtr, ok := opts.Config["dtr"]
	if !ok {
		return nil, fmt.Errorf("dtr config not found")
	}
	host, _ := dtr["host"].(string)
	user, _ := dtr["user"].(string)
	pass, _ := dtr["pass"].(string)

	return &Suite{
		rand:   opts.Rand,
		logger: opts.Logger,
		client: client.New(client.Opts{
			Host: host,
			User: user,
			Pass: pass,
		}),
	}, nil
}

func (s *Suite) GetCommand(cmd string) (types.TestCommand, error) {
	return modules.GetCommand(s, cmd)
}

func (s *Suite) CreateUser(a types.TestArgs) (types.TestResult, error) {
	user := map[string]interface{}{
		"name":     a.String("username"),
		"password": a.String("password"),
		"isActive": true,
		"isAdmin":  a.Bool("isadmin"),
	}

	_, err := s.client.Do("POST", "/enzi/v0/accounts", user)
	return nil, err
}

func (s *Suite) CreateRandomUser(a types.TestArgs) (types.TestResult, error) {
	user := map[string]interface{}{
		"name":     util.RandomString(s.rand, 10),
		"password": a.String("password"),
		"isActive": true,
		"isAdmin":  a.Bool("isadmin"),
	}

	s.logger.WithField("data", user).Info("creating user")

	return s.client.Do("POST", "/enzi/v0/accounts", user)
}

func (s *Suite) CreateRepo(a types.TestArgs) (types.TestResult, error) {
	data := map[string]interface{}{
		"name":       util.RandomString(s.rand, 10),
		"visibility": "public",
	}
	return s.client.Do("POST", "/api/v0/repositories/"+a.String("namespace"), data)
}

func (s *Suite) CreateUserAndRepo(a types.TestArgs) (types.TestResult, error) {
	user, _ := s.CreateRandomUser(types.TestArgs{
		"password": "test",
	})

	return s.CreateRepo(types.TestArgs{
		"namespace": user["name"],
	})
}
