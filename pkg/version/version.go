package version

import (
	"fmt"
	"runtime"
)

var (
	gitVersion   = "dev"
	gitCommit    string
	gitTreeState string
	buildDate    string

	info = Info{
		GitVersion:   gitVersion,
		GitCommit:    gitCommit,
		GitTreeState: gitTreeState,
		BuildDate:    buildDate,
		GoVersion:    runtime.Version(),
		Compiler:     runtime.Compiler,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
)

func GetInfo() Info {
	return info
}

func GitVersion() string {
	return gitVersion
}
