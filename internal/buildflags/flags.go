package buildflags

import "fmt"

var (
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
	BuildVersion string = "N/A"
)

// PrintBuildInformation prints the build version, build date, and build commit.
func PrintBuildInformation() {
	fmt.Printf("Build version: %s\nBuild date: %s\nBuild commit: %s\n", BuildVersion, BuildDate, BuildCommit)
}
