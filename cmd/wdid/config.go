package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/asdine/storm"
	"gitlab.com/josler/wdid/core"
)

type ConfigStore struct {
	Type string
	File string
}

type ConfigAuto struct {
	Type     string
	Key      string
	Username string
}

func (ca ConfigAuto) AutoType() string {
	return ca.Type
}

func (ca ConfigAuto) AuthKey() string {
	return ca.Key
}

func (ca ConfigAuto) AutoUsername() string {
	return ca.Username
}

type Config struct {
	Store ConfigStore
	Auto  []ConfigAuto
}

var defaultConfig = `
[store]
type = "bolt"
file = "~/.config/wdid/wdid.db"

#[[auto]]
#type = "github"
#key = "apikey"
#username = "username"
`

func loadConfig() (*Config, error) {
	filename := filepath.Join(configDir(), "config.toml")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.MkdirAll(configDir(), os.ModePerm)

		fmt.Println("setting up config...")
		file, err := os.Create(filename)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		_, err = file.WriteString(defaultConfig)
		if err != nil {
			fmt.Println(err)
		}

		file.Sync()
		file.Close()
	}

	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}
	var conf Config
	if _, err := toml.DecodeReader(file, &conf); err != nil {
		return nil, err
	}
	if conf.Auto == nil {
		conf.Auto = []ConfigAuto{}
	}
	return &conf, nil
}

func createStore(conf *Config) (core.Store, error) {
	switch conf.Store.Type {
	case "bolt":
		db, err := storm.Open(parseConfigPath(conf.Store.File))
		if err != nil {
			return nil, err
		}
		return core.NewBoltStore(db), nil
	default:
		return nil, errors.New("store not specified correctly")
	}
}

func parseConfigPath(filepath string) string {
	pathCmd := exec.Command("sh", "-c", fmt.Sprintf("echo %s", filepath))
	out, _ := pathCmd.CombinedOutput()
	pathCmd.Run()
	return strings.Trim(string(out), "\n")
}

func configDir() string {
	return filepath.Join(homeDir(), ".config", "wdid")
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "defaultuser"
	}
	return usr.HomeDir
}
