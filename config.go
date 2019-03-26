package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// NewConfig is used to read config file
func NewConfig(configFile string, configType string) *Config {
	var config Config

	switch configType {
	case `yaml`:
		InitiateYaml(&config, configFile)
		break
	case `yaml-base64`:
		InitiateYamlBase64(&config, configFile)
		break
	case `env`:
		InitiateFromEnv(&config)
		break
	default:
		InitiateYaml(&config, configFile)
	}

	return &config
}

// InitiateYaml initiates config from yaml file
func InitiateYaml(config *Config, filepath string) {
	configdata, errReadFile := ioutil.ReadFile(filepath)
	if errReadFile != nil {
		log.Fatalf(errReadFile.Error())
	}

	errParseYaml := yaml.Unmarshal(configdata, &config)
	if errParseYaml != nil {
		log.Fatalf(errParseYaml.Error())
	}
}

// InitiateYamlBase64 initiates config from environment variable and the value is in base64
func InitiateYamlBase64(config *Config, envkey string) {
	tempConfigDataInBase64 := os.Getenv(envkey)
	configdata, err := base64.StdEncoding.DecodeString(tempConfigDataInBase64)
	if err != nil {
		log.Fatalf(err.Error())
	}

	errParseYaml := yaml.Unmarshal([]byte(configdata), &config)
	if errParseYaml != nil {
		log.Fatalf(errParseYaml.Error())
	}
}

// InitiateFromEnv initiates config from environment variable
// This is applicable for single task only
func InitiateFromEnv(config *Config) {
	config.Source.Type = os.Getenv("SOURCE_TYPE")
	config.Source.Host = os.Getenv("SOURCE_HOST")
	config.Source.Port = os.Getenv("SOURCE_PORT")
	config.Source.Username = os.Getenv("SOURCE_USERNAME")
	config.Source.Password = os.Getenv("SOURCE_PASSWORD")

	config.Target.Type = os.Getenv("TARGET_TYPE")
	config.Target.Host = os.Getenv("TARGET_HOST")

	targetHeader := os.Getenv("TARGET_HEADER")
	targetHeaders := strings.Split(targetHeader, `;`)
	for _, item := range targetHeaders {
		tempItem := strings.Split(item, `:`)
		if len(tempItem) > 1 {
			header := make(map[string]string)
			header[`key`] = tempItem[0]
			header[`value`] = tempItem[1]

			config.Target.Header = append(config.Target.Header, header)
		}
	}

	timeout := "5"
	if os.Getenv("TIMEOUT") != "" {
		timeout = os.Getenv("TIMEOUT")
	}
	uploadTimeout, errUploadTimeout := strconv.ParseInt(timeout, 10, 64)
	if errUploadTimeout != nil {
		uploadTimeout = 5
	}

	config.Target.Timeout = uploadTimeout

	targetUploadParam := os.Getenv("TARGET_UPLOAD_PARAM")
	targetUploadParams := strings.Split(targetUploadParam, `;`)
	for _, item := range targetUploadParams {
		tempItem := strings.Split(item, `:`)
		if len(tempItem) > 1 {
			param := make(map[string]string)
			param[`key`] = tempItem[0]
			param[`value`] = tempItem[1]

			config.Target.Upload = append(config.Target.Upload, param)
		}
	}

	cronEvery, errCronEvery := strconv.ParseUint(os.Getenv("CRON_EVERY"), 10, 64)
	if errCronEvery != nil {
		cronEvery = 0
	}

	filePrefixIndex, errFilePrefixIndex := strconv.ParseInt(os.Getenv("TASK_FILE_PREFIX_INDEX"), 10, 64)
	if errFilePrefixIndex != nil {
		filePrefixIndex = 0
	}

	cron := Cron{
		Name:        os.Getenv("CRON_NAME"),
		Type:        os.Getenv("CRON_TYPE"),
		SpecificDay: os.Getenv("CRON_SPECIFIC_DAY"),
		At:          os.Getenv("CRON_AT"),
		Every:       cronEvery,
		Task: CronTask{
			SourceFolder:        os.Getenv("TASK_FOLDER"),
			File:                os.Getenv("TASK_FILE"),
			FilePrefix:          os.Getenv("TASK_FILE_PREFIX"),
			FilePrefixDelimiter: os.Getenv("TASK_FILE_PREFIX_DELIMITER"),
			FilePrefixIndex:     filePrefixIndex,
		},
	}

	config.Cron = append(config.Cron, cron)
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
	Type    string              `yaml:"type"`
	Host    string              `yaml:"host"`
	Header  []map[string]string `yaml:"header"`
	Upload  []map[string]string `yaml:"upload"`
	Timeout int64               `yaml:"timeout"`
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
