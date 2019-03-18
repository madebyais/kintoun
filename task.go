package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jasonlvhit/gocron"
)

// NewTask returns available method in Task schema
func NewTask(config *Config, client *Client) *Task {
	return &Task{
		config:          config,
		client:          client,
		lastFileModTime: make(map[string]time.Time),
		lastFileUpload:  make(map[string]string),
	}
}

// Task represents
type Task struct {
	config *Config
	client *Client

	filenameToDownload string
	lastFileModTime    map[string]time.Time
	lastFileUpload     map[string]string
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

		job.Do(t.SendFile(item))
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

// SendFile is used to upload file into target destination
func (t *Task) SendFile(crondata Cron) func() {
	return func() {
		clientSession := t.client.InitSftp()
		defer clientSession.Close()

		Logf("Job name=%s\n", crondata.Name)
		if crondata.Task.FilePrefix != "" {
			sourceFiles, errSourceFiles := clientSession.ReadDir(crondata.Task.SourceFolder)
			if errSourceFiles != nil {
				Logf("Failed to list directory dir=%s error=%s\n", crondata.Task.SourceFolder, errSourceFiles.Error())
			}

			for _, item := range sourceFiles {
				isMatch, _ := regexp.MatchString(crondata.Task.FilePrefix, item.Name())
				isYearMatch := item.ModTime().Year() == time.Now().Year()
				isMonthMatch := item.ModTime().Month() == time.Now().Month()
				isDayMatch := item.ModTime().Day() == time.Now().Day()

				prefixCodes := strings.Split(item.Name(), crondata.Task.FilePrefixDelimiter)
				prefixCode := prefixCodes[crondata.Task.FilePrefixIndex]
				isFileLatestUpdate := item.ModTime().After(t.lastFileModTime[prefixCode])
				isPrevFileDifferent := t.lastFileUpload[prefixCode] != item.Name()

				if !isPrevFileDifferent {
					t.filenameToDownload = ""
					continue
				}

				if isMatch && isYearMatch && isMonthMatch && isDayMatch && isFileLatestUpdate && isPrevFileDifferent {
					t.lastFileModTime[prefixCode] = item.ModTime()
					t.lastFileUpload[prefixCode] = item.Name()
					t.filenameToDownload = item.Name()
				}
			}
		} else {
			t.filenameToDownload = crondata.Task.File
		}

		if t.filenameToDownload == "" {
			Log("No new file need to be downloaded")
			Log("----------------------------------")
			return
		}

		Logf("Downloading file=%s/%s ...\n", crondata.Task.SourceFolder, t.filenameToDownload)

		filepath := crondata.Task.SourceFolder + "/" + t.filenameToDownload
		sourceFile, errSourceFile := clientSession.Open(filepath)
		if errSourceFile != nil {
			Logf("Failed to download filepath=%s error=%s\n", filepath, errSourceFile.Error())
			Log("----------------------------------")
			return
		}
		defer sourceFile.Close()

		tempfile := fmt.Sprintf("./%s", t.filenameToDownload)
		destinationFile, errCreateDestFile := os.Create(tempfile)
		if errCreateDestFile != nil {
			Logf("Failed to create temp file when downloading error=%s\n", errCreateDestFile.Error())
			Log("----------------------------------")
			return
		}
		defer destinationFile.Close()

		_, errCopySourceToDest := io.Copy(destinationFile, sourceFile)
		if errCopySourceToDest != nil {
			Logf("Failed to download error=%s\n", errCopySourceToDest.Error())
			Log("----------------------------------")
			return
		}
		destinationFile.Sync()
		Logf("File=%s/%s Temp=%s has been downloaded succesfully ...\n", crondata.Task.SourceFolder, t.filenameToDownload, tempfile)

		t.UploadFile(tempfile)
	}
}

// UploadFile will upload from download temp file into target destination
func (t *Task) UploadFile(tempfilepath string) {
	Logf("Uploading file=%s ...\n", tempfilepath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, uploadItem := range t.config.Target.Upload {
		if uploadItem["key"] == uploadItem["value"] {
			file, err := os.Open(tempfilepath)
			if err != nil {
				Logf("Failed to upload file=%s error=%s\n", tempfilepath, err.Error())
				Log("----------------------------------")
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
		Logf("Failed to upload file=%s ...\n", tempfilepath)
		Log("----------------------------------")
		return
	}

	for _, header := range t.config.Target.Header {
		req.Header.Set(header["key"], header["value"])
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpclient := &http.Client{}
	resp, err := httpclient.Do(req)
	if err != nil {
		Logf("Failed to receive response when uploading file=%s ...\n", tempfilepath)
		Log("----------------------------------")
		return
	}

	if resp.StatusCode != 200 {
		Logf("Failed to upload file got status_code=%s\n", strconv.Itoa(resp.StatusCode))
		Log("Retrying file upload in 5s ...")
		Log("----------------------------------")
		time.Sleep(5 * time.Second)
		t.UploadFile(tempfilepath)
		return
	}

	Log("Removing temp file ...")
	_ = os.Remove(tempfilepath)

	Log("Uploaded successfully")
	Log("----------------------------------")
}
