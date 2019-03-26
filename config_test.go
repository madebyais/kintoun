package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfigWithEnvVariables(t *testing.T) {
	os.Setenv("SOURCE_TYPE", "ftps")
	os.Setenv("SOURCE_HOST", "0.0.0.0")
	os.Setenv("SOURCE_PORT", "23")
	os.Setenv("SOURCE_USERNAME", "jetbrains")
	os.Setenv("SOURCE_PASSWORD", "jetbrains")
	os.Setenv("TARGET_TYPE", "http")
	os.Setenv("TARGET_HOST", "https://55413b45.ngrok.io/reconciliation/upload-file")
	os.Setenv("TARGET_HEADER", "Authorization:Basic 12345")
	os.Setenv("TARGET_UPLOAD_PARAM", "file:file;channel:CIMB;statement:withdrawal")
	os.Setenv("TIMEOUT", "60")
	os.Setenv("CRON_NAME", "get-sample-txt")
	os.Setenv("CRON_EVERY", "2")
	os.Setenv("CRON_TYPE", "second")
	os.Setenv("CRON_SPECIFIC_DAY", "None")
	os.Setenv("CRON_AT", "14:32")
	os.Setenv("TASK_FOLDER", "/share/dumps")
	os.Setenv("TASK_FILE_PREFIX", `\d*_\d*_\d*.\d*.\d*.csv$`)
	os.Setenv("TASK_FILE_PREFIX_DELIMITER", ".")
	os.Setenv("TASK_FILE_PREFIX_INDEX", "1")
	os.Setenv("TASK_FILE", "sample.txt")

	config := NewConfig("", "env")

	assert.Equal(t, "ftps", config.Source.Type)
	assert.Equal(t, "http", config.Target.Type)
	assert.Equal(t, "get-sample-txt", config.Cron[0].Name)
}
