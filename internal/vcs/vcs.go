package vcs

import (
	"runtime/debug"
)

func Version() string {
	//Use debug.ReadBuildInfo to retrive a debug.BuildInfo struct. If this is available,
	//the okor value will be true, and we return the pseudo-version contained in the Main.Version field
	bi, ok := debug.ReadBuildInfo()
	if ok {
		return bi.Main.Version
	}
	return ""
}
