package selfupdate

import (
	"fmt"
	"runtime"
	"strings"
)

const binaryName = "stackit-acl"

func FindAsset(release *Release) (*Asset, error) {
	suffix := fmt.Sprintf("_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
	for i := range release.Assets {
		name := strings.ToLower(release.Assets[i].Name)
		if strings.HasPrefix(name, binaryName+"_") && strings.HasSuffix(name, suffix) {
			return &release.Assets[i], nil
		}
	}
	return nil, fmt.Errorf("no release asset found for %s/%s", runtime.GOOS, runtime.GOARCH)
}

func FindAssetByName(release *Release, name string) *Asset {
	for i := range release.Assets {
		if release.Assets[i].Name == name {
			return &release.Assets[i]
		}
	}
	return nil
}
