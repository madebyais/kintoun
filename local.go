package main

import (
	"io/ioutil"
	"regexp"
	"strings"
	"time"
)

// LocalFolder client
type LocalFolder struct {
	dirpath            string
	filenameToDownload string
	lastFileModTime    map[string]time.Time
	lastFileUpload     map[string]string
}

// NewLocalFolder initiates local folder client
func NewLocalFolder(dirpath string) Interface {
	return &LocalFolder{
		dirpath:         dirpath,
		lastFileModTime: LastFileModTime,
		lastFileUpload:  LastFileUpload,
	}
}

// ReaddirSourceFolder is used to read files in a local directory
func (l *LocalFolder) ReaddirSourceFolder(crontdata Cron) error {
	files, err := ioutil.ReadDir(l.dirpath)
	if err != nil {
		return err
	}

	for _, item := range files {
		if crontdata.Task.FilePrefix != "" {
			isMatch, _ := regexp.MatchString(crontdata.Task.FilePrefix, item.Name())
			isYearMatch := item.ModTime().Year() == time.Now().Year()
			isMonthMatch := item.ModTime().Month() == time.Now().Month()
			isDayMatch := item.ModTime().Day() == time.Now().Day()

			prefixCodes := strings.Split(item.Name(), crontdata.Task.FilePrefixDelimiter)
			prefixCode := prefixCodes[crontdata.Task.FilePrefixIndex]
			isFileLatestUpdate := item.ModTime().After(l.lastFileModTime[prefixCode])
			isPrevFileDifferent := l.lastFileUpload[prefixCode] != item.Name()

			if !isPrevFileDifferent {
				l.SetFilenameToDownload("")
				continue
			}

			if isMatch && isYearMatch && isMonthMatch && isDayMatch && isFileLatestUpdate && isPrevFileDifferent {
				l.lastFileModTime[prefixCode] = item.ModTime()
				l.lastFileUpload[prefixCode] = item.Name()
				l.SetFilenameToDownload(item.Name())
			}
		} else {
			l.SetFilenameToDownload(crontdata.Task.File)
		}
	}

	return nil
}

// SetFilenameToDownload is used to set a filename to download as temp file
func (l *LocalFolder) SetFilenameToDownload(filename string) {
	l.filenameToDownload = l.dirpath + "/" + filename
}

// GetFilenameToDownload is used to get a filename to download as temp file
func (l *LocalFolder) GetFilenameToDownload() string {
	return l.filenameToDownload
}

// DownloadTempFile will not be used since the file is already in local directory
func (l *LocalFolder) DownloadTempFile(filepath string) error {
	return nil
}

// Close for local folder is do nothing
func (l *LocalFolder) Close() {}
