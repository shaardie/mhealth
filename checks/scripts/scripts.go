package scripts

import (
	"os/exec"

	"github.com/shaardie/mhealth/checks/interfaces"
)

type Config struct {
	GivenName string `yaml:"name"`
	Path      string `yaml:"path"`
}

type scriptCheck Config

func Init(cfg Config) (interfaces.Check, error) {
	c := scriptCheck(cfg)
	return &c, nil
}

func (c *scriptCheck) Name() string {
	return c.GivenName
}

func (c *scriptCheck) Type() string {
	return "scripts"
}

func (c *scriptCheck) Run() error {
	return exec.Command(c.Path).Run()
}
