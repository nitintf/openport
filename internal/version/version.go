package version

// These are set at build time via -ldflags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

func Full() string {
	if Version == "dev" {
		return "dev"
	}
	return Version
}
