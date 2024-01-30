package config

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"io/ioutil"
	"log"
)

// Config 配置文件
type Config struct {
	ServerIp   string   `yaml:"serverIp"`
	ServerPort string   `yaml:"serverPort"`
	ClientName string   `yaml:"clientName"`
	Plugins    []string `yaml:"plugins"`
}

func NewConfig() *Config {
	return &Config{}
}

func loadConfigFile(path string) []byte {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println(err)
		return nil
	}
	return b
}

func LoadConfig(path string) (*Config, error) {
	config := NewConfig()
	if b := loadConfigFile(path); b != nil {

		err := yaml.Unmarshal(b, config)
		if err != nil {
			return nil, err
		}
		return config, err
	} else {
		return nil, fmt.Errorf("load config file error")
	}

}
