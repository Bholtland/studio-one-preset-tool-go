package config

import (
	"os"
	"path"
	"regexp"
)

type in struct {
	Path     string
	FileName string
	Full     string
}

type out struct {
	Path string
}

type temp struct {
	Path                   string
	SongContentsPath       string
	PresetConstructionPath string
}

type Config struct {
	In                in
	Out               out
	Temp              temp
	RemoveExistingOut bool
}

func New(inPath string, outPath string, removeExistingOut bool) *Config {
	re := regexp.MustCompile(`(.*)\/(.*\.song)`)
	pathInMatch := re.FindStringSubmatch(inPath)

	if pathInMatch[1] == "" && pathInMatch[2] == "" {
		panic("No valid in path set")
	}

	if outPath == "" {
		panic("No valid out path set")
	}

	tempPath := path.Join(os.TempDir(), "studio-one-preset-tool")

	return &Config{
		In: in{
			Path:     pathInMatch[1],
			FileName: pathInMatch[2],
			Full:     inPath,
		},
		Out: out{
			Path: outPath,
		},
		Temp: temp{
			Path:                   tempPath,
			SongContentsPath:       path.Join(tempPath, "song-contents"),
			PresetConstructionPath: path.Join(tempPath, "preset-construction"),
		},
		RemoveExistingOut: removeExistingOut,
	}
}
