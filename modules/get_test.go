package modules

import (
	"reflect"
	"testing"

	"github.com/docker/integreat/errors"
	"github.com/docker/integreat/types"
)

type ExampleSuite struct {
	private string
}

func (e *ExampleSuite) GetCommand(string) (types.TestCommand, error) {
	return nil, nil
}

func (e *ExampleSuite) DoSomething(t types.TestArgs) (types.TestResult, error) {
	return types.TestResult{"foo": e.private}, nil
}

func TestGet(t *testing.T) {
	suite := &ExampleSuite{
		private: "bar",
	}

	tests := []struct {
		Command    string
		Error      error
		TestResult types.TestResult
		TestError  error
	}{
		{"DoSomething", nil, types.TestResult{"foo": "bar"}, nil},
		{"invalid", errors.ErrCommandNotFound, nil, nil},
	}

	for _, item := range tests {
		f, err := Get(suite, item.Command)
		if err != item.Error {
			t.Fatal("unexpected error")
		}

		if item.Error != nil {
			continue
		}

		// Assert calling functions from Get work as expected
		result, err := f(types.TestArgs{})
		if !reflect.DeepEqual(result, item.TestResult) {
			t.Fatal("unexpected result")
		}

		if err != item.TestError {
			t.Fatal("unexpected test error")
		}
	}
}
