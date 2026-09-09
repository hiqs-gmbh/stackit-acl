package selfupdate

import "strings"

func Normalize(version string) string {
	return strings.TrimPrefix(version, "v")
}

func Tag(version string) string {
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}
