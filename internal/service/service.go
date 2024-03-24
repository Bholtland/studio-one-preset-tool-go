package service

import "bholtland/studio-one-preset-tool-go/internal/config"

type element struct {
	Type string 
	Name string
	Attributes map[string]string
	Elements []element
}

type Service struct {
	cfg *config.Config
}

type Preset struct {
	DeviceClassID  string
	DeviceBaseName string
	DeviceCategory string
	DeviceSubCategory string
	DeviceName string
	DeviceUID string
	TrackID string
	Name string
}

func New(cfg *config.Config) *Service {
	return &Service{
		cfg: cfg,
	}
}


func (s *Service) CreatePreset(preset Preset) error {
	// Make temp directory

	// Move default files to temp directory

  // get meta info and write it to file

	// get preset parts and write it to file

	// get directory structure

	// compress and write to directry (from directory structure)
}

func (s *Service) mapPresetParts(preset Preset) *element {
}

func (s *Service) mapMetaInfo(preset Preset) *element {
	return &element{
		Elements: []element{
			{
				Type: "element",
				Name: "MetaInformation",
				Elements: []element{
					{
						Type: "element",
						Name: "Attribute",
						Attributes: map[string]string{
							"id": "Class:ID",
							"value": preset.DeviceClassID,
						},
					},
				},
			},
	 },
	}
}

func (s *Service) getDirectoryStructure(preset Preset) *element {
	
}