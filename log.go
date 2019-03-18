package main

import "fmt"

// Log is a normal println log
func Log(text string) {
	fmt.Println("[ KINTOUN ] " + text)
}

// Logf is a normal printf log
func Logf(text string, params ...interface{}) {
	fmt.Printf("[ KINTOUN ] "+text, params...)
}
