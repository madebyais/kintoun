package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

// NewConfig is used to read config file
func NewConfig(filepath string, configType string) *Config {
	var config Config

	if configType == `yaml` {
		configdata, errReadFile := ioutil.ReadFile(filepath)
		if errReadFile != nil {
			log.Fatalf(errReadFile.Error())
		}

		errParseYaml := yaml.Unmarshal(configdata, &config)
		if errParseYaml != nil {
			log.Fatalf(errParseYaml.Error())
		}
	} else {
		tempConfigDataInBase64 := os.Getenv(filepath)
		configdata, err := base64.StdEncoding.DecodeString(tempConfigDataInBase64)
		if err != nil {
			log.Fatalf(err.Error())
		}

		errParseYaml := yaml.Unmarshal([]byte(configdata), &config)
		if errParseYaml != nil {
			log.Fatalf(errParseYaml.Error())
		}
	}

	return &config
}

// Config represents the config file for go-upload
type Config struct {
	Source Source `yaml:"source" json:"source"`
	Target Target `yaml:"target" json:"target"`
	Cron   []Cron `yaml:"cron" json:"cron"`
}

// Source represents parameter used for get data from source data
type Source struct {
	Type     string `yaml:"type"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Folder   string `yaml:"folder"`
}

// Target represents parameter used for submit data to target data
type Target struct {
	Type   string              `yaml:"type"`
	Host   string              `yaml:"host"`
	Header []map[string]string `yaml:"header"`
	Upload []map[string]string `yaml:"upload"`
}

// Cron represents parameter used for schedule task
type Cron struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	SpecificDay string   `yaml:"specific_day"`
	At          string   `yaml:"at"`
	Every       uint64   `yaml:"every"`
	Task        CronTask `yaml:"task"`
}

// CronTask specifies source folder and the file that want to be uploaded
type CronTask struct {
	SourceFolder        string `yaml:"folder"`
	File                string `yaml:"file"`
	FilePrefix          string `yaml:"file_prefix"`
	FilePrefixDelimiter string `yaml:"file_prefix_delimiter"`
	FilePrefixIndex     int64  `yaml:"file_prefix_index"`
}
