package integreat

import (
	"fmt"
	"io/ioutil"

	"github.com/docker/integreat/config"
	"github.com/docker/integreat/modules"
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

	return &Suite{
		config:  config,
		logger:  opts.Logger,
		modules: map[string]types.Module{},
	}, nil
}

// Suite represents the entire suite of tests defined by the YAML file to run
type Suite struct {
	logger *logrus.Logger
	config *config.Configuration

	modules map[string]types.Module
}

func (s *Suite) Run() error {
	err := s.initModules()
	if err != nil {
		s.logger.WithError(err).Error("error initializing modules")
		return err
	}

	s.logger.WithField("config", s.config).Info("starting tests")
	return nil
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

		config, _ := s.config.Config[name]
		s.modules[name], err = creator(config)
		if err != nil {
			return err
		}
	}

	return nil
}
