package main

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/webguerilla/ftps"
)

// FTPS client
type FTPS struct {
	ftpsclient         *ftps.FTPS
	filenameToDownload string
	lastFileModTime    map[string]time.Time
	lastFileUpload     map[string]string
}

// NewFTPS initiates FTPS client
func NewFTPS(host, port, username, password string) Interface {
	ftpsClient := new(ftps.FTPS)
	ftpsClient.TLSConfig.InsecureSkipVerify = true
	ftpsClient.Debug = false

	hostPort, _ := strconv.Atoi(port)
	errConnect := ftpsClient.Connect(host, hostPort)
	if errConnect != nil {
		panic(errConnect)
	}

	errLogin := ftpsClient.Login(username, password)
	if errLogin != nil {
		panic(errLogin)
	}

	return &FTPS{
		ftpsclient:      ftpsClient,
		lastFileModTime: LastFileModTime,
		lastFileUpload:  LastFileUpload,
	}
}

// ReaddirSourceFolder is used to read files in a dir
func (f *FTPS) ReaddirSourceFolder(crondata Cron) error {
	if crondata.Task.FilePrefix != "" {
		errChangeWorkingDir := f.ftpsclient.ChangeWorkingDirectory(crondata.Task.SourceFolder)
		if errChangeWorkingDir != nil {
			return errChangeWorkingDir
		}

		entries, errListWorkingDir := f.ftpsclient.List()
		if errListWorkingDir != nil {
			return errListWorkingDir
		}

		for _, item := range entries {
			isMatch, _ := regexp.MatchString(crondata.Task.FilePrefix, item.Name)
			isYearMatch := item.Time.Year() == time.Now().Year()
			isMonthMatch := item.Time.Month() == time.Now().Month()
			isDayMatch := item.Time.Day() == time.Now().Day()

			prefixCodes := strings.Split(item.Name, crondata.Task.FilePrefixDelimiter)
			prefixCode := prefixCodes[crondata.Task.FilePrefixIndex]
			isFileLatestUpdate := item.Time.After(f.lastFileModTime[prefixCode])
			isPrevFileDifferent := f.lastFileUpload[prefixCode] != item.Name

			if !isPrevFileDifferent {
				f.SetFilenameToDownload("")
				continue
			}

			if isMatch && isYearMatch && isMonthMatch && isDayMatch && isFileLatestUpdate && isPrevFileDifferent {
				f.lastFileModTime[prefixCode] = item.Time
				f.lastFileUpload[prefixCode] = item.Name
				f.SetFilenameToDownload(item.Name)
			}
		}
	} else {
		f.SetFilenameToDownload(crondata.Task.File)
	}

	return nil
}

// SetFilenameToDownload is used to set a filename to download as temp file
func (f *FTPS) SetFilenameToDownload(filename string) {
	f.filenameToDownload = filename
}

// GetFilenameToDownload is used to get a filename to download as temp file
func (f *FTPS) GetFilenameToDownload() string {
	return f.filenameToDownload
}

// DownloadTempFile will download the file
func (f *FTPS) DownloadTempFile(filepath string) error {
	Logf("Downloading file=%s ...\n", filepath)

	sourceFile, errSourceFile := f.ftpsclient.RetrieveFileData(filepath)
	if errSourceFile != nil {
		return errSourceFile
	}

	tempfile := fmt.Sprintf("./%s", f.GetFilenameToDownload())
	errWriteFile := ioutil.WriteFile(tempfile, sourceFile, 0644)
	if errWriteFile != nil {
		return errWriteFile
	}

	Log("File has been downloaded succesfully ...")
	return nil
}

// Close is used to close a connection
func (f *FTPS) Close() {
	f.ftpsclient.Quit()
}
