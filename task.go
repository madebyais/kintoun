package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
)

// Task represents task that will be executed
type Task struct {
	config *Config
}

// NewTask returns a task object
func NewTask(config *Config) *Task {
	return &Task{
		config: config,
	}
}

// Start will start running the job in background
func (t *Task) Start() {
	t.Register()

	_, time := gocron.NextRun()
	Log("Service started at " + time.String())
	Log("----------------------------------")

	<-gocron.Start()
}

// Register is used to register new cron task
func (t *Task) Register() {
	Log("Register job started")
	for _, item := range t.config.Cron {
		job := gocron.Every(item.Every)
		job = t.getJobType(job, item.Every, item.Type)

		if item.At != "" {
			job = job.At(item.At)
		}

		job.Do(t.Exec(item))
		Logf("Job name=%s every=%d type=%s specific_day=%s at=%s is registered ...\n", item.Name, item.Every, item.Type, item.SpecificDay, item.At)
	}

	Log("Done registering jobs")
}

func (t *Task) getJobType(job *gocron.Job, every uint64, defaultType string) *gocron.Job {
	if every > 1 {
		defaultType = defaultType + "s"
	}

	switch defaultType {
	case "second":
		job = job.Second()
		break
	case "seconds":
		job = job.Seconds()
		break
	case "minute":
		job = job.Minute()
		break
	case "minutes":
		job = job.Minutes()
		break
	case "hour":
		job = job.Hour()
		break
	case "hours":
		job = job.Hours()
		break
	case "days":
		job = job.Days()
		break
	case "day":
		job = job.Day()
		break
	default:
		job = job.Day()
	}

	return job
}

// Exec will execute the job based on submitted config
// Following are the steps for exec method:
//
// - Read files in source folder/dir
// - If there's a new file to download, then set the filename
// - Download the file as temporary file
// - Upload to target destionation
// - Remove temporary file
func (t *Task) Exec(crondata Cron) func() {
	return func() {
		Logf("Job name=%s\n", crondata.Name)

		clientType := strings.ToLower(t.config.Source.Type)
		clientSession := InitiateFTPClient(clientType, t.config)
		defer clientSession.Close()

		folderPath := crondata.Task.SourceFolder
		errReaddirSourceFolder := clientSession.ReaddirSourceFolder(crondata)
		if errReaddirSourceFolder != nil {
			Logf("Failed to list directory dir=%s error=%s\n", folderPath, errReaddirSourceFolder.Error())
			Log("----------------------------------")
			return
		}

		filename := clientSession.GetFilenameToDownload()
		if filename == "" {
			Log("No new file need to be downloaded")
			Log("----------------------------------")
			return
		}

		filepath := folderPath + `/` + filename
		errDownloadTempFile := clientSession.DownloadTempFile(filepath)
		if errDownloadTempFile != nil {
			Logf("Failed to download filepath=%s error=%s\n", filepath, errDownloadTempFile.Error())
			Log("----------------------------------")
			return
		}

		t.Upload(filename)
	}
}

// Upload is used to uplad download temp file to destination
func (t *Task) Upload(tempfilepath string) {
	Logf("Uploading file=%s ...\n", tempfilepath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, uploadItem := range t.config.Target.Upload {
		if uploadItem["key"] == uploadItem["value"] {
			file, err := os.Open(tempfilepath)
			if err != nil {
				Logf("Failed to upload file=%s error=%s\n", tempfilepath, err.Error())
				Log("----------------------------------")
				return
			}
			defer file.Close()

			part, err := writer.CreateFormFile(uploadItem["key"], filepath.Base(tempfilepath))
			if err != nil {
				Logf("Failed to upload file=%s error=%s\n", tempfilepath, err.Error())
				Log("----------------------------------")
				return
			}
			_, _ = io.Copy(part, file)
		} else {
			writer.WriteField(uploadItem["key"], uploadItem["value"])
		}
	}

	errWriterClose := writer.Close()
	if errWriterClose != nil {
		Logf("Failed to upload file=%s error=%s\n", tempfilepath, errWriterClose.Error())
		Log("----------------------------------")
		return
	}

	req, err := http.NewRequest("POST", t.config.Target.Host, body)
	if err != nil {
		Logf("Failed to upload file=%s error=%s...\n", tempfilepath, err.Error())
		Log("----------------------------------")
		return
	}

	for _, header := range t.config.Target.Header {
		req.Header.Set(header["key"], header["value"])
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpclient := &http.Client{Timeout: time.Duration(t.config.Target.Timeout) * time.Second}
	resp, err := httpclient.Do(req)
	if err != nil {
		Logf("Failed to receive response when uploading file=%s error=%s\n", tempfilepath, err.Error())
		Log("Retrying file upload in 5s ...")
		Log("----------------------------------")
		time.Sleep(5 * time.Second)
		t.Upload(tempfilepath)
		return
	}

	if resp.StatusCode != 200 {
		Logf("Failed to upload file got status_code=%s\n", strconv.Itoa(resp.StatusCode))
		Log("Retrying file upload in 5s ...")
		Log("----------------------------------")
		time.Sleep(5 * time.Second)
		t.Upload(tempfilepath)
		return
	}

	Log("File has been uploaded successfully")

	Log("Removing temp file ...")
	_ = os.Remove(tempfilepath)

	Logf("Job is done\n")
	Log("----------------------------------")
}
