package version

import "fmt"

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func String() string {
	return fmt.Sprintf("wblue %s (commit=%s, built=%s)", Version, GitCommit, BuildTime)
}
