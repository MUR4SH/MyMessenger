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
		Address          string `yaml:"address"`
		Database         string `yaml:"database"`
		Messages         string `yaml:"messages"`
		Users            string `yaml:"users"`
		Chats            string `yaml:"chats"`
		Files            string `yaml:"files"`
		ChatSettings     string `yaml:"chat_settings"`
		ChatsArray       string `yaml:"chats_array"`
		PersonalSettings string `yaml:"personal_settings"`
	}
	API struct {
		Port string `yaml:"port"`
	} `yaml:"web"`
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

	dbInterface := databaseInterface.New(
		config.Database.Address,
		config.Database.Database,
		config.Database.Messages,
		config.Database.Users,
		config.Database.Chats,
		config.Database.Files,
		config.Database.ChatSettings,
		config.Database.ChatsArray,
		config.Database.PersonalSettings,
	)

	serverAndHandlers.InitServer(config.API.Port, &dbInterface)

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
