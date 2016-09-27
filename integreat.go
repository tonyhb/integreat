package integreat

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/docker/integreat/config"
	"github.com/docker/integreat/modules"
	_ "github.com/docker/integreat/modules/dtr"
	_ "github.com/docker/integreat/modules/registry"
	"github.com/docker/integreat/types"

	"github.com/Sirupsen/logrus"
)

type Opts struct {
	Logger *logrus.Logger

	// ConfigPath is the location of the config yaml file for the integreat
	// test suite
	ConfigPath string
}

// New returns a new test suite to run
func New(opts Opts) (*Suite, error) {
	byt, err := ioutil.ReadFile(opts.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration: %s", err)
	}

	config, err := config.Parse(byt)
	if err != nil {
		return nil, fmt.Errorf("error reading configuration: %s", err)
	}

	seed := config.Base.Seed
	if seed == 0 {
		seed = time.Now().Unix()
	}

	return &Suite{
		logger:  opts.Logger,
		rand:    rand.New(rand.NewSource(seed)),
		config:  config,
		modules: map[string]types.Module{},
		results: map[string][]types.TestResult{},
	}, nil
}

// Suite represents the entire suite of tests defined by the YAML file to run
type Suite struct {
	logger *logrus.Logger
	rand   *rand.Rand
	config *types.Configuration

	modules map[string]types.Module

	results map[string][]types.TestResult
}

func (s *Suite) Run() error {
	err := s.initModules()
	if err != nil {
		s.logger.WithError(err).Error("error initializing modules")
		return err
	}

	args := types.TestArgs{}

	for _, test := range s.config.Tests {
		s.logger.WithFields(logrus.Fields{
			"id":      test.Id,
			"name":    test.Name,
			"command": test.Command,
			"args":    test.Args,
			"repeat":  test.Repeat,
		}).Info("running command")

		cmd, err := s.resolveCommand(test.Command)
		if err != nil {
			s.logger.WithError(err).Error("error resolving command")
			return err
		}

		if test.Repeat == 0 {
			test.Repeat = 1
		}

		for i := 1; i <= test.Repeat; i++ {
			if test.Args == nil {
				test.Args = types.TestArgs{}
			}
			for k, v := range args {
				test.Args[k] = v
			}
			result, err := cmd(test.Args)
			if err != nil {
				s.logger.WithError(err).Error("error running command")
				return err
			}

			if _, ok := args[test.Id]; ok {
				args[test.Id] = append(args[test.Id].([]types.TestResult), result)
			} else {
				args[test.Id] = []types.TestResult{result}
			}
		}

	}

	return nil
}

func (s *Suite) resolveCommand(cmd string) (types.TestCommand, error) {
	// each command is in the format of "module::FuncName"
	parts := strings.SplitN(cmd, "::", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid command '%s'", cmd)
	}
	module, ok := s.modules[parts[0]]
	if !ok {
		return nil, fmt.Errorf("unknown module '%s'", parts[0])
	}

	return module.GetCommand(parts[1])
}

// initModules attempts to construct each module suite with config options
// defined in the YAML file.
//
// This errors if any config suite is not found or if any config suite throws
// an error during initialization, usually due to incorrect configuration
func (s *Suite) initModules() error {
	for _, name := range s.config.Modules {
		s.logger.WithField("module", name).Debug("initiating module")

		creator, err := modules.GetModule(name)
		if err != nil {
			return err
		}

		s.modules[name], err = creator(types.ModuleOpts{
			Config: s.config.Config,
			Logger: s.logger,
			Rand:   s.rand,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
