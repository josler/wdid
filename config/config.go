package config

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
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
	Store     ConfigStore
	Auto      []ConfigAuto
	ColorTags bool `toml:"color_tags"`
}

var defaultConfig = `
color_tags = false

[store]
type = "bolt"
file = "~/.config/wdid/wdid.db"

#[[auto]]
#type = "github"
#key = "apikey"
#username = "username"
`

func Load() (*Config, error) {
	filename := filepath.Join(ConfigDir(), "config.toml")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		os.MkdirAll(ConfigDir(), os.ModePerm)

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
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var conf Config
	if _, err := toml.DecodeReader(file, &conf); err != nil {
		return nil, err
	}
	if conf.Auto == nil {
		conf.Auto = []ConfigAuto{}
	}
	return &conf, nil
}

func ConfigDir() string {
	return filepath.Join(homeDir(), ".config", "wdid")
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "defaultuser"
	}
	return usr.HomeDir
}

func (store ConfigStore) Filepath() string {
	pathCmd := exec.Command("sh", "-c", fmt.Sprintf("echo %s", store.File))
	out, _ := pathCmd.CombinedOutput()
	pathCmd.Run()
	return strings.Trim(string(out), "\n")
}
