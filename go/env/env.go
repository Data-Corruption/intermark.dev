package env

import "os"

const (
	IM_ADDRESS        = "IM_ADDRESS"
	IM_PAGE_CACHE_MB  = "IM_PAGE_CACHE_MB"
	IM_ASSET_CACHE_MB = "IM_ASSET_CACHE_MB"
	IM_LOG_LEVEL      = "IM_LOG_LEVEL"
	IM_UPDATE_SECRET  = "IM_UPDATE_SECRET"

	// Timeouts in minutes

	IM_GIT_M  = "IM_GIT_M"
	IM_LFS_M  = "IM_LFS_M"
	IM_TAIL_M = "IM_TAIL_M"
	IM_LUNR_M = "IM_LUNR_M"
)

var defaults = map[string]string{
	IM_ADDRESS:        ":9292",
	IM_PAGE_CACHE_MB:  "1024", // 1GB
	IM_ASSET_CACHE_MB: "1024", // 1GB
	IM_LOG_LEVEL:      "warn",
	IM_UPDATE_SECRET:  "",
	IM_GIT_M:          "5",
	IM_LFS_M:          "5",
	IM_TAIL_M:         "1",
	IM_LUNR_M:         "1",
}

func Get(key string) string {
	if envValue, exists := os.LookupEnv(key); exists {
		return envValue
	}
	return defaults[key]
}
