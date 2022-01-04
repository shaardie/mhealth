package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"

	"github.com/shaardie/mhealth/api"
	"github.com/shaardie/mhealth/checks"
	"github.com/shaardie/mhealth/storage"
)

type config struct {
	Api      api.Config          `yaml:"api"`
	Database storage.Config      `yaml:"storage"`
	Checks   checks.Config       `yaml:"checks"`
	Actions  map[string][][]byte `yaml:"actions"`
}

func readConfig(filename string) (*config, error) {
	c := &config{}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return c, fmt.Errorf("unable to read config file %v, %w", filename, err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return c, fmt.Errorf("unable to unmarshal config file %v, %w", filename, err)
	}
	return c, err
}

func mainWithErrros() error {
	cfg, err := readConfig("config.yaml")
	if err != nil {
		return fmt.Errorf("failed to process config, %w", err)
	}
	log.Printf("config: %+v", *cfg)

	db, err := storage.Init(cfg.Database)
	if err != nil {
		return fmt.Errorf("failed to initialize storage, %w", err)
	}

	cm, err := checks.Init(cfg.Checks, db)
	if err != nil {
		return fmt.Errorf("failed to initialize check manager, %w", err)
	}

	s, err := api.Init(cfg.Api, db)
	if err != nil {
		return fmt.Errorf("failed to initialize api, %w", err)
	}

	go s.Run()
	cm.Run()

	return nil

}

func main() {
	err := mainWithErrros()
	if err != nil {
		log.Fatalln(err)
	}
}
