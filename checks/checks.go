package checks

import (
	"fmt"
	"log"
	"time"

	"github.com/shaardie/mhealth/checks/commands"
	"github.com/shaardie/mhealth/checks/interfaces"
	"github.com/shaardie/mhealth/checks/scripts"
	"github.com/shaardie/mhealth/storage"
)

type Config struct {
	Interval int               `yaml:"interval"`
	Scripts  []scripts.Config  `yaml:"scripts"`
	Commands []commands.Config `yaml:"commands"`
}

type CheckManager struct {
	cfg    Config
	db     storage.DB
	checks []interfaces.Check
}

func Init(cfg Config, db storage.DB) (*CheckManager, error) {
	cm := &CheckManager{
		cfg:    cfg,
		db:     db,
		checks: make([]interfaces.Check, 0),
	}

	for _, scriptCfg := range cfg.Scripts {
		c, err := scripts.Init(scriptCfg)
		if err != nil {
			return cm, fmt.Errorf("failed to init script check %+v, %w", scriptCfg, err)
		}
		cm.checks = append(cm.checks, c)
		log.Printf("Added check %v/%v", c.Type(), c.Name())
	}

	for _, commandCfg := range cfg.Commands {
		c, err := commands.Init(commandCfg)
		if err != nil {
			return cm, fmt.Errorf("failed to init command check %+v, %w", commandCfg, err)
		}
		cm.checks = append(cm.checks, c)
		log.Printf("Added check %v/%v", c.Type(), c.Name())
	}

	return cm, nil
}

func (cm CheckManager) Run() {
	for {
		err := cm.SingleRun()
		if err != nil {
			log.Printf("Failed to run checks, %v", err)
		}
		time.Sleep(time.Duration(cm.cfg.Interval) * time.Second)
	}
}

func (cm CheckManager) SingleRun() error {
	for _, check := range cm.checks {
		log.Printf("Run check %v/%v", check.Type(), check.Name())
		err := check.Run()
		failed := 0
		if err != nil {
			log.Printf("Check %v/%v failed, %v", check.Type(), check.Name(), err)
			failed = 1
		}
		_, err = cm.db.CreateOrUpdateCheck.Exec(check.Type(), check.Name(), failed, failed)
		if err != nil {
			return fmt.Errorf("unable to write check %v/%v to database, %w", check.Type(), check.Name(), err)
		}
	}

	return nil
}
