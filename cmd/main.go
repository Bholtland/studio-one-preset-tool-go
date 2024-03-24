package main

import (
	"archive/zip"
	"bholtland/studio-one-preset-tool-go/internal/config"
	"bholtland/studio-one-preset-tool-go/internal/song"
	"fmt"
	"os"
	"path"

	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
  		Name:   "greet",
  		Usage:  "say a greeting",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "in-path",
					Value: "C:/Users/beren/OneDrive/Documents/Studio One/Songs/Instrument Exploration/Instrument Exploration.song",
					Usage: "The path to the song file",
					EnvVar: "IN_PATH",
				},
				&cli.StringFlag{
					Name:  "out-path",
					Value: "C:/Users/beren/OneDrive/Documents/Studio One Autogenerated Presets",
					Usage: "The path to the output directory",
					EnvVar: "OUT_PATH",
				},
				&cli.BoolFlag{
					Name:  "remove-existing",
					Usage: "Whether to remove existing files in the output directory",
					EnvVar: "REMOVE_EXISTING",
				},
			},
			Action: func(c *cli.Context) error {
				cfg := config.New(c.String("in-path"), c.String("out-path"), c.Bool("remove-existing"))
				return run(c, cfg)
			},
	}  
	

	err := app.Run(os.Args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}


func run(ctx *cli.Context, cfg *config.Config) error {
	defer cleanup(cfg)

	err := extractProject(cfg)
	if err != nil {
		return fmt.Errorf("Error extracting project: %s", err)
	}

	songContents, err := song.Parse()
	if err != nil {
		return fmt.Errorf("Error parsing song: %s", err)
	}
}

func cleanup(cfg *config.Config) error {}
	

func extractProject(config *config.Config) error {
	read, err := zip.OpenReader(config.In.Full)
	if err != nil { 
		return err 
	}
	defer read.Close()

	for _, file := range read.File {
		if file.Mode().IsDir() {
			continue 
		}

		open, err := file.Open()
		if err != nil {
			return err 
		}

		name := path.Join(fmt.Sprintf("%s/%s", config.TempDir, "songContents"))
		os.MkdirAll(path.Dir(name), os.ModeDir)
		create, err := os.Create(name)
		if err != nil { 
			return err 
		}
		defer create.Close()
		 
		create.ReadFrom(open)
	}
	return nil
}