package writer

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	"bholtland/studio-one-preset-tool-go/internal/reader"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/saracen/fastzip"
	"io"
	"os"
	"path/filepath"
)

type metaAttribute struct {
	ID    string `xml:"id,attr"`
	Value string `xml:"value,attr"`
}

type metaInfo struct {
	// TODO: Is this right?
	XMLName         xml.Name `xml:"MetaInformation"`
	MetaInformation struct {
		Attributes []metaAttribute `xml:"Attribute"`
	} `xml:"MetaInformation"`
}

type presetParts struct {
	PresetParts struct {
		PresetPart []struct {
			Attributes []metaAttribute `xml:"Attribute"`
		} `xml:"PresetPart"`
	} `xml:"PresetParts"`
}

type Service struct {
	cfg *config.Config
	ctx context.Context
}

func NewService(cfg *config.Config, ctx context.Context) *Service {
	return &Service{
		cfg: cfg,
		ctx: ctx,
	}
}

func (s *Service) CreatePresets(presetMap *reader.PresetMap) error {
	// Make concurrent
	for _, preset := range *presetMap {
		if err := s.createPreset(preset); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) createPreset(preset *reader.PresetMapEntry) error {
	// Create preset dir
	if err := os.MkdirAll(s.cfg.In.Path, os.ModeTemporary); err != nil {
		return err
	}

	// Copy raw preset file
	if err := s.copyFile(
		fmt.Sprintf("%s/songContents/%s", s.cfg.In.Path, preset.Path),
		fmt.Sprintf("%s/temp/%s/%s", preset.SongID, preset.FileName),
	); err != nil {
		return err
	}

	metaInfoContent := s.buildMetaInfo(preset)
	if err := s.writeXML(metaInfoContent, fmt.Sprintf("%s/temp/%s/metainfo.xml", s.cfg.In.Path, preset.SongID)); err != nil {
		return err
	}

	presetPartsContent := s.buildPresetParts(preset)
	if err := s.writeXML(presetPartsContent, fmt.Sprintf("%s/temp/%s/presetparts.xml", s.cfg.In.Path, preset.SongID)); err != nil {
		return err
	}

	if err := s.compressPreset(preset); err != nil {
		return err
	}

	return nil
}

func (s *Service) buildMetaInfo(preset *reader.PresetMapEntry) *metaInfo {

	return &metaInfo{
		MetaInformation: struct {
			Attributes []metaAttribute `xml:"Attribute"`
		}{
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
		},
	}
}

func (s *Service) buildPresetParts(preset *reader.PresetMapEntry) *presetParts {
	return &presetParts{
		PresetParts: struct {
			PresetPart []struct {
				Attributes []metaAttribute `xml:"Attribute"`
			} `xml:"PresetPart"`
		}{
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
		},
	}
}

func (s *Service) compressPreset(preset *reader.PresetMapEntry) error {
	if err := os.MkdirAll(fmt.Sprintf("%s/%s", s.cfg.Out.Path, preset.Path), os.ModeDir); err != nil {
		return err
	}

	// Create archive file
	w, err := os.Create(fmt.Sprintf("%s/%s/%s.instrument", s.cfg.Out.Path, preset.Path, preset.FileName))
	if err != nil {
		panic(err)
	}
	defer w.Close()

	source := fmt.Sprintf("%s/temp/%s", s.cfg.In.Path, preset.SongID)

	// Create new Archiver
	a, err := fastzip.NewArchiver(w, source)
	if err != nil {
		panic(err)
	}
	defer a.Close()

	// Walk directory, adding the files we want to add
	files := make(map[string]os.FileInfo)
	err = filepath.Walk(source, func(pathname string, info os.FileInfo, err error) error {
		files[pathname] = info
		return nil
	})

	// Archive
	if err = a.Archive(s.ctx, files); err != nil {
		panic(err)
	}

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
