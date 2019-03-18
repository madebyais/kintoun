package main

import (
	"flag"
)

func main() {
	var configFile string
	var configType string

	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file path")
	flag.StringVar(&configType, "config-type", "yaml", "Configuration type: yaml, yaml-base64")

	flag.Parse()

	config := NewConfig(configFile, configType)
	client := NewClient(config)

	task := NewTask(config, client)
	task.Start()
}
