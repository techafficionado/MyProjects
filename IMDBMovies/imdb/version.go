package main

import "fmt"

const (
	// PRINCIPAL version represent a significant / incompatible API change.
	PRINCIPAL = 1

	// MAJOR version represent the addition of functionality in a backward-compatible manner.
	MAJOR = 0

	// MINOR version represent a bug fix(es) in a backward-compatible manner.
	MINOR = 0
)

// Version returns the version number of the application.
func Version() string {
	return fmt.Sprintf("%d.%d.%d", PRINCIPAL, MAJOR, MINOR)
}
