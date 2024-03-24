package config

import (
	"os"
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

type Config struct {
	In                in
	Out               out
	TempDir           string
	RemoveExistingOut bool
}

func New(inPath string, outPath string, removeExistingOut bool) *Config {
	re := regexp.MustCompile(`(.*)\/(.*\.song)`)
	pathInMatch := re.FindStringSubmatch(inPath)

	if (pathInMatch[1] == "" && pathInMatch[2] == "") {
		panic("No valid in path set")
	}

	if (outPath == "") {
		panic("No valid out path set")
	}

	return &Config{
		In: in{
			Path:     pathInMatch[1],
			FileName: pathInMatch[2],
			Full:     inPath,
		},
		Out: out{
			Path: outPath,
		},
		TempDir:     os.TempDir(),
		RemoveExistingOut: removeExistingOut,
	}
}