package song

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	xml "encoding/xml"
	"os"
)

type Song struct {
	SongData             SongData
	AudioSynthFolderData AudioSynthFolderData
	MusicTrackDeviceData MusicTrackDeviceData
}

type SongData struct {
}

type AudioSynthFolderData struct {
}

type MusicTrackDeviceData struct {
}

func Parse() (Song, error)  {}

func (s *Song) getSongData(cfg config.Config) (*SongData, error) {
	rawXml, err := os.ReadFile(cfg.In.Full)
	if err != nil {
		return nil, err
	}

	var data SongData

	data := xml.Unmarshal(rawXml, )

}