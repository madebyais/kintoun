package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// Interface is FTP client interfaces
type Interface interface {
	ReaddirSourceFolder(crontdata Cron) error
	SetFilenameToDownload(filename string)
	GetFilenameToDownload() string
	DownloadTempFile(filepath string) error
	Close()
}

// InitiateFTPClient will initiates ftp client based on client type, whether it is a FTP/s or SFTP
// By default it will use SFTP
func InitiateFTPClient(clientType string, config *Config) Interface {
	host := config.Source.Host
	port := config.Source.Port
	username := config.Source.Username
	password := config.Source.Password
	dirpath := config.Source.Folder

	var clientSession Interface

	switch clientType {
	case `sftp`:
		clientSession = NewSFTP(host, port, username, password)
		break
	case `ftps`:
		clientSession = NewFTPS(host, port, username, password)
		break
	case `local`:
		clientSession = NewLocalFolder(dirpath)
		break
	default:
		clientSession = NewSFTP(host, port, username, password)
	}

	return clientSession
}

// Upload is used to uplad download temp file to destination
func Upload(config *Config, tempfilepath string) error {
	Logf("Uploading file=%s ...\n", tempfilepath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, uploadItem := range config.Target.Upload {
		if uploadItem["key"] == uploadItem["value"] {
			file, err := os.Open(tempfilepath)
			if err != nil {
				return err
			}
			defer file.Close()

			part, err := writer.CreateFormFile(uploadItem["key"], filepath.Base(tempfilepath))
			if err != nil {
				return err
			}
			_, _ = io.Copy(part, file)
		} else {
			writer.WriteField(uploadItem["key"], uploadItem["value"])
		}
	}

	errWriterClose := writer.Close()
	if errWriterClose != nil {
		return errWriterClose
	}

	req, err := http.NewRequest("POST", config.Target.Host, body)
	if err != nil {
		return err
	}

	for _, header := range config.Target.Header {
		req.Header.Set(header["key"], header["value"])
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	httpclient := &http.Client{}
	resp, err := httpclient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return Upload(config, tempfilepath)
	}

	return nil
}
