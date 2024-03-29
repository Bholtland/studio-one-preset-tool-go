package writer

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	"bholtland/studio-one-preset-tool-go/internal/reader"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/saracen/fastzip"
	"io"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type metaAttribute struct {
	ID    string `xml:"id,attr"`
	Value string `xml:"value,attr"`
}

type metaInfo struct {
	XMLName    xml.Name        `xml:"MetaInformation"`
	Attributes []metaAttribute `xml:"Attribute"`
}

type presetParts struct {
	XMLName    xml.Name `xml:"PresetParts"`
	PresetPart []struct {
		Attributes []metaAttribute `xml:"Attribute"`
	} `xml:"PresetPart"`
}

type Service struct {
	cfg    *config.Config
	ctx    context.Context
	logger *slog.Logger
}

func NewService(cfg *config.Config, ctx context.Context, logger *slog.Logger) *Service {
	return &Service{
		cfg:    cfg,
		ctx:    ctx,
		logger: logger,
	}
}

func (s *Service) CreatePresets(presetMap *reader.PresetMap) error {
	if err := os.RemoveAll(s.cfg.Out.Path); err != nil {
		return err
	}

	// Create a buffered channel for errors
	errs := make(chan error, runtime.NumCPU())

	// Create a WaitGroup
	var wg sync.WaitGroup

	if presetMap == nil {
		return errors.New("PresetMap is nil")
	}

	// Loop over presets
	for _, preset := range *presetMap {
		// Increment the WaitGroup counter
		wg.Add(1)

		// Start a new goroutine
		go func(p reader.PresetMapEntry) {
			err := s.createPreset(&p)

			// If there was an error, send it on the errs channel
			if err != nil {
				errs <- err
			}

			// Decrement the WaitGroup counter
			wg.Done()
		}(*preset)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Close the errs channel
	close(errs)

	// Check if there were any errors
	for err := range errs {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) createPreset(preset *reader.PresetMapEntry) error {
	// Create preset dir
	if err := os.MkdirAll(path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID), os.ModeTemporary); err != nil {
		return err
	}

	// Copy raw preset file
	if err := s.copyFile(
		path.Join(s.cfg.Temp.SongContentsPath, "Presets", "Synths", preset.FileName),
		path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID, preset.FileName),
	); err != nil {
		return err
	}

	metaInfoContent := s.buildMetaInfo(preset)
	if err := s.writeXML(metaInfoContent, path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID, "metainfo.xml")); err != nil {
		return err
	}

	presetPartsContent := s.buildPresetParts(preset)
	if err := s.writeXML(presetPartsContent, path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID, "presetparts.xml")); err != nil {
		return err
	}

	if err := s.compressPreset(preset); err != nil {
		return err
	}

	return nil
}

func (s *Service) buildMetaInfo(preset *reader.PresetMapEntry) *metaInfo {
	return &metaInfo{
		Attributes: []metaAttribute{
			{
				ID:    "Class:ID",
				Value: preset.DeviceClassID,
			},
			{
				ID:    "Class:Name",
				Value: preset.DeviceBaseName,
			},
			{
				ID:    "Class:Category",
				Value: preset.DeviceCategory,
			},
			{
				ID:    "Class:SubCategory",
				Value: preset.DeviceSubCategory,
			},
			{
				ID:    "DeviceSlot:deviceName",
				Value: preset.DeviceName,
			},
			{
				ID:    "DeviceSlot:deviceUID",
				Value: preset.DeviceUID,
			},
			{
				ID:    "DeviceSlot:slotUID",
				Value: preset.TrackID,
			},
			{
				ID:    "Document:Title",
				Value: preset.Name,
			},
			{
				ID:    "Document:Creator",
				Value: "Studio One Preset Tool",
			},
			{
				ID:    "Document:Generator",
				Value: "Studio One Preset Tool",
			},
		},
	}
}

func (s *Service) buildPresetParts(preset *reader.PresetMapEntry) *presetParts {
	return &presetParts{
		PresetPart: []struct {
			Attributes []metaAttribute `xml:"Attribute"`
		}{
			{
				Attributes: []metaAttribute{
					{
						ID:    "Class:ID",
						Value: preset.DeviceClassID,
					},
					{
						ID:    "Class:Name",
						Value: preset.DeviceBaseName,
					},
					{
						ID:    "Class:Category",
						Value: preset.DeviceCategory,
					},
					{
						ID:    "Class:SubCategory",
						Value: preset.DeviceSubCategory,
					},
					{
						ID:    "DeviceSlot:deviceName",
						Value: preset.DeviceName,
					},
					{
						ID:    "DeviceSlot:deviceUID",
						Value: preset.DeviceUID,
					},
					{
						ID:    "DeviceSlot:slotUID",
						Value: preset.TrackID,
					},
					{
						ID:    "AudioSynth:IsMainPreset",
						Value: "1",
					},
					{
						ID:    "Preset:DataFile",
						Value: preset.FileName,
					},
				},
			},
		},
	}
}

func (s *Service) compressPreset(preset *reader.PresetMapEntry) error {
	if err := os.MkdirAll(path.Join(s.cfg.Out.Path, preset.Path), os.ModeDir); err != nil {
		return err
	}

	normalizedName := strings.ReplaceAll(preset.Name, "\"", " inch")
	// Create archive file
	w, err := os.Create(path.Join(s.cfg.Out.Path, preset.Path, fmt.Sprintf("%s.instrument", normalizedName)))
	if err != nil {
		return err
	}
	defer w.Close()

	sourcePath := path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID)

	// Create new Archiver
	a, err := fastzip.NewArchiver(w, sourcePath)
	if err != nil {
		return err
	}
	defer a.Close()

	// Walk directory, adding the files we want to add
	files := make(map[string]os.FileInfo)
	err = filepath.Walk(sourcePath, func(pathname string, info os.FileInfo, err error) error {
		files[pathname] = info
		return nil
	})

	// Archive
	if err = a.Archive(s.ctx, files); err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("Created %s", fmt.Sprintf("%s/%s.instrument", preset.Path, normalizedName)))

	return nil
}

func (s *Service) copyFile(src string, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	err = destFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) writeXML(content interface{}, path string) error {
	// Create XML file
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	if _, err = file.WriteString(xml.Header); err != nil {
		return err
	}

	// Write XML content
	encoder := xml.NewEncoder(file)
	encoder.Indent("", "  ")
	err = encoder.Encode(content)
	if err != nil {
		return err
	}

	// Close file
	err = file.Close()
	if err != nil {
		return err
	}

	return nil
}
