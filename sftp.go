package main

import (
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTP client
type SFTP struct {
	sftpclient         *sftp.Client
	filenameToDownload []string
	lastFileModTime    map[string]time.Time
	lastFileUpload     map[string]string
}

// NewSFTP initiates SFTP client
func NewSFTP(host, port, username, password string) Interface {
	sshconfig := &ssh.ClientConfig{
		User: username,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	hostAddr := fmt.Sprintf("%s:%s", host, port)

	sshclient, errSSHClient := ssh.Dial("tcp", hostAddr, sshconfig)
	if errSSHClient != nil {
		panic(errSSHClient)
	}

	sftpclient, errSftpClient := sftp.NewClient(sshclient)
	if errSftpClient != nil {
		panic(errSftpClient)
	}

	return &SFTP{
		sftpclient:      sftpclient,
		lastFileModTime: LastFileModTime,
		lastFileUpload:  LastFileUpload,
	}
}

// ReaddirSourceFolder is used to read files in a dir
func (s *SFTP) ReaddirSourceFolder(crondata Cron) error {
	fileToDownload := make([]string, 0)

	if crondata.Task.FilePrefix != "" {
		sourceFiles, errSourceFiles := s.sftpclient.ReadDir(crondata.Task.SourceFolder)
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
			isFileLatestUpdate := item.ModTime().After(s.lastFileModTime[prefixCode])
			isPrevFileDifferent := s.lastFileUpload[prefixCode] != item.Name()

			if !isPrevFileDifferent {
				continue
			}

			if isMatch && isYearMatch && isMonthMatch && isDayMatch && isFileLatestUpdate && isPrevFileDifferent {
				s.lastFileModTime[prefixCode] = item.ModTime()
				s.lastFileUpload[prefixCode] = item.Name()
				fileToDownload = append(fileToDownload, item.Name())
			}
		}
	} else {
		fileToDownload = append(fileToDownload, crondata.Task.File)
	}

	s.SetFilenameToDownload(fileToDownload)

	return nil
}

// SetFilenameToDownload is used to set a filename to download as temp file
func (s *SFTP) SetFilenameToDownload(filename []string) {
	s.filenameToDownload = filename
}

// GetFilenameToDownload is used to get a filename to download as temp file
func (s *SFTP) GetFilenameToDownload() []string {
	return s.filenameToDownload
}

// DownloadTempFile will download the file
func (s *SFTP) DownloadTempFile(filepath string) error {
	Logf("Downloading file=%s ...\n", filepath)

	sourceFile, errSourceFile := s.sftpclient.Open(filepath)
	if errSourceFile != nil {
		return errSourceFile
	}
	defer sourceFile.Close()

	tempfile := fmt.Sprintf("./%s", s.GetFilenameToDownload())
	destinationFile, errCreateDestFile := os.Create(tempfile)
	if errCreateDestFile != nil {
		return errCreateDestFile
	}
	defer destinationFile.Close()

	_, errCopySourceToDest := io.Copy(destinationFile, sourceFile)
	if errCopySourceToDest != nil {
		return errCopySourceToDest
	}
	destinationFile.Sync()
	Log("File has been downloaded succesfully ...")

	return nil
}

// Close is used to close a connection
func (s *SFTP) Close() {
	s.sftpclient.Close()
}
