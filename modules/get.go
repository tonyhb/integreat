package modules

import (
	"fmt"
	"reflect"

	"github.com/docker/integreat/errors"
	"github.com/docker/integreat/types"
)

func GetModule(name string) (types.ModuleCreator, error) {
	creator, ok := modules[name]
	if !ok {
		return nil, fmt.Errorf("module '%s' is not registered", name)
	}
	return creator, nil
}

// Get returns a command from a given suite
func GetCommand(m types.Module, cmd string) (types.TestCommand, error) {
	val := reflect.ValueOf(m)
	if val.CanAddr() {
		val = val.Addr()
	}

	f := val.MethodByName(cmd)
	if f.IsValid() {
		return types.TestCommand(f.Interface().(func(types.TestArgs) (types.TestResult, error))), nil
	}

	return nil, errors.ErrCommandNotFound
}
