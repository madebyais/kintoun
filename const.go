package main

import "time"

var LastFileModTime map[string]time.Time = make(map[string]time.Time)
var LastFileUpload map[string]string = make(map[string]string)
