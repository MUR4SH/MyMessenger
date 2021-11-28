package main

import (
	"os"
	"sync"

	"github.com/MUR4SH/MyMessenger/databaseInterface"
	"github.com/MUR4SH/MyMessenger/serverAndHandlers"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Database struct {
		Address  string `yaml:"address"`
		Username string `yaml:"username"`
		Password string `yaml:"password"`
		Database string `yaml:"database"`
	}
	API struct {
		Port string `yaml:"port"`
	} `yaml:"api"`
}

func main() {

	confFile, err := os.Open("config.yaml")
	if err != nil {
		panic(err)
	}

	var config Config
	err = yaml.NewDecoder(confFile).Decode(&config)
	if err != nil {
		panic(err)
	}
	confFile.Close()

	dbInterface := databaseInterface.New(config.Database.Address, config.Database.Username, config.Database.Password, config.Database.Database)

	serverAndHandlers.InitServer(config.API.Port, &dbInterface)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
