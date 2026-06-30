package version

import (
	"os"
	"time"
)

var (
	App       = "tree-api"
	Version   = "0.1.0"
	Commit    = "unknown"
	BuildTime = "unknown"

	runtimeBuildTime = time.Now().Format(time.RFC3339)
)

type Info struct {
	App         string `json:"app"`
	Version     string `json:"version"`
	Commit      string `json:"commit"`
	BuildTime   string `json:"buildTime"`
	Environment string `json:"environment"`
}

func Current(environment string) Info {
	return Info{
		App:         valueFromEnv("TREE_APP_NAME", App),
		Version:     valueFromEnv("TREE_APP_VERSION", Version),
		Commit:      valueFromEnv("TREE_GIT_COMMIT", Commit),
		BuildTime:   valueFromEnv("TREE_BUILD_TIME", normalizeBuildTime(BuildTime)),
		Environment: valueFromEnv("TREE_BUILD_ENV", environment),
	}
}

func valueFromEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func normalizeBuildTime(value string) string {
	if value != "" && value != "unknown" {
		return value
	}
	return runtimeBuildTime
}
