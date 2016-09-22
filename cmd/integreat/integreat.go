package main

import (
	"fmt"
	"os"

	"github.com/docker/integreat"

	"github.com/Sirupsen/logrus"
)

func main() {
	if len(os.Args) == 1 {
		fmt.Println("usage: `integreat /path/to/yaml.yml`")
		os.Exit(1)
	}

	suite, err := integreat.New(integreat.Opts{
		ConfigPath: os.Args[1],
		Logger:     logrus.StandardLogger(),
	})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = suite.Run()
	if err != nil {
		fmt.Printf("testing error\n")
		os.Exit(1)
	}
}
