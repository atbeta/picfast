package version

const defaultGitHubURL = "https://github.com/atbeta/picfast"

var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func Info() map[string]string {
	return map[string]string{
		"version":    Version,
		"commit":     Commit,
		"build_time": BuildTime,
		"github_url": defaultGitHubURL,
	}
}

func DefaultGitHubURL() string {
	return defaultGitHubURL
}
