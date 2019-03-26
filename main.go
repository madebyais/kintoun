package main

import (
	"flag"
)

func main() {
	var configFile string
	var configType string

	flag.StringVar(&configFile, "config", "config.yaml", "Configuration file path")
	flag.StringVar(&configType, "config-type", "yaml", "Configuration type: yaml, yaml-base64, env")

	flag.Parse()

	config := NewConfig(configFile, configType)

	task := NewTask(config)
	task.Start()
}
