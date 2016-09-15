package modules

import (
	"github.com/docker/integreat/errors"
	"github.com/docker/integreat/types"
)

var (
	modules map[string]types.ModuleCreator
)

func init() {
	modules = make(map[string]types.ModuleCreator)
}

func Register(name string, f types.ModuleCreator) error {
	if _, ok := modules[name]; ok {
		return errors.ErrModuleRegistered
	}
	modules[name] = f
	return nil
}
