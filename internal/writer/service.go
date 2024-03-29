package writer

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	"bholtland/studio-one-preset-tool-go/internal/file"
	"bholtland/studio-one-preset-tool-go/internal/reader"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
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
	if err := file.Copy(
		path.Join(s.cfg.Temp.SongContentsPath, "Presets", "Synths", preset.FileName),
		path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID, preset.FileName),
	); err != nil {
		return err
	}

	metaInfoContent := s.buildMetaInfo(preset)
	if err := file.WriteXML(metaInfoContent, path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID, "metainfo.xml")); err != nil {
		return err
	}

	presetPartsContent := s.buildPresetParts(preset)
	if err := file.WriteXML(presetPartsContent, path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID, "presetparts.xml")); err != nil {
		return err
	}

	normalizedName := strings.ReplaceAll(preset.Name, "\"", " inch")
	if err := file.Compress(
		s.ctx, path.Join(s.cfg.Temp.PresetConstructionPath, preset.SongID),
		path.Join(s.cfg.Out.Path, preset.Path),
		fmt.Sprintf("%s.instrument", normalizedName),
	); err != nil {
		return err
	}

	s.logger.Info(fmt.Sprintf("Created %s", fmt.Sprintf("%s/%s.instrument", preset.Path, normalizedName)))

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
