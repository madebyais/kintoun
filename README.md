KINTOUN
-------

![Build Status](https://travis-ci.org/madebyais/kintoun.svg?branch=master)

#### Background

There are several projects that I've been working on need for an automation in uploading files from one place to another place.

#### Supports

Currently, it only supports from `FTP`, `FTPS`, `SFTP`, to `HTTP REST API`.

#### How-to

Copy the sample config file below.

```
source:
  type: sftp
  host: 0.0.0.0
  port: 22
  username: foo
  password: pass

target:
  type: http
  host: http://www.kintoun.com/upload-file
  header:
    - key: Authorization
      value: Basic 12345
  upload:
    - key: file 
      value: file
    - key: channel
      value: CIMB

cron:
  - name: get-sample-txt
    every: 5 
    type: second 
    specific_day: None
    at: '14:32'
    task: 
      folder: /upload
      file: sample.txt
```

Find below for explanation.

```
source:
  type: sftp 
  host: 0.0.0.0
  port: 22
  username: foo
  password: pass
```
`source` is the source data. Currently, KINTOUN only supports SFTP.

`source.type` is can be set to `sftp`, `ftp`, or `ftps` 

`source.host` is the host of the sftp server

`source.port` is the port of the sftp server

`source.username` is the username to access ftp server

`source.password` is the password to access ftp server


```
target:
  type: http
  host: http://www.kintoun.com/upload-file
  header:
    - key: Authorization
      value: Basic 12345
  upload:
    - key: file 
      value: file
    - key: channel
      value: some_3rd_party
```
`target` is the destination where the data will be sent. Currently, KINTOUN only supports HTTP.

`target.type` is set to `http`

`target.host` is the url for the destination server

`target.header` is the header that need to be sent along with the request to destination server

`target.upload` contains the list of form to be sent to the destination server. If the `key` and `value` are same, then KINTOUN will set this as the file object in the multipart form


```
cron:
  - name: get-sample-txt
    every: 5 
    type: second 
    specific_day: None
    at: '14:32'
    task: 
      folder: /upload
      file: sample.txt
```
`cron` contains the list of job that will run

`cron.name` is the name of the job

`cron.type` is the type of time, such as second, minute, hour, day

`cron.specific_day` is specific to a day, it should be set to None if `cron.type` is second, minute, or hour. Available list: None, Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday

`cron.at` this only specific to day, about what time job will be run e.g. `15:30`

`cron.task.folder` is the source folder

`cron.task.file` is the source file

#### LICENSE

MIT
