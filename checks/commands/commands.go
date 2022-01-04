package commands

import (
	"os/exec"

	"github.com/shaardie/mhealth/checks/interfaces"
)

type Config struct {
	GivenName string   `yaml:"name"`
	Command   string   `yaml:"command"`
	Args      []string `yaml:"args"`
}

type commandCheck Config

func Init(cfg Config) (interfaces.Check, error) {
	c := commandCheck(cfg)
	return &c, nil
}

func (c *commandCheck) Name() string {
	return c.GivenName
}

func (c *commandCheck) Type() string {
	return "scripts"
}

func (c *commandCheck) Run() error {
	return exec.Command(c.Command, c.Args...).Run()
}
