package main

import "os"

func main() {
	// This should be flagged by osexit analyzer
	os.Exit(1) // want "os.Exit should not be called in main function of main package"
	
	// Nested call should also be detected
	if true {
		os.Exit(0) // want "os.Exit should not be called in main function of main package"
	}
}

func helper() {
	// This is OK - os.Exit in non-main function is allowed
	os.Exit(1)
}

func anotherFunc() {
	// This is also OK
	defer os.Exit(1)
}
