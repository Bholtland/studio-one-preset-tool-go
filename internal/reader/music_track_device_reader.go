package reader

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	"bholtland/studio-one-preset-tool-go/internal/file"
	"encoding/xml"
	"log/slog"
	"path"
	"regexp"
)

type MusicTrackDeviceXML struct {
	XMLName    xml.Name `xml:"MusicTrackDevice"`
	Attributes struct {
		ChannelGroup struct {
			MusicTrackChannel []struct {
				Connection []struct {
					XID      string `xml:"id,attr"`
					ObjectID string `xml:"objectID,attr"`
				} `xml:"Connection"`
				UID []struct {
					XID string `xml:"id,attr"`
					UID string `xml:"uid,attr"`
				} `xml:"UID"`
			} `xml:"MusicTrackChannel"`
		} `xml:"ChannelGroup"`
	} `xml:"Attributes"`
}

type MusicTrackDeviceMap map[string]*MusicTrackDeviceMapEntry

type MusicTrackDeviceMapEntry struct {
	MusicTrackDeviceID string
	SongID             string
}

type MusicTrackDeviceReader struct {
	cfg *config.Config
}

func NewMusicTrackDeviceReader(cfg *config.Config) *MusicTrackDeviceReader {
	return &MusicTrackDeviceReader{
		cfg: cfg,
	}
}

func (s *MusicTrackDeviceReader) GetMap() (MusicTrackDeviceMap, error) {
	xml, err := file.ReadXML[MusicTrackDeviceXML](path.Join(s.cfg.Temp.SongContentsPath, "Devices", "musictrackdevice.xml"))
	if err != nil {
		return nil, err
	}

	return s.BuildMusicTrackDeviceMap(xml), nil
}

func (s *MusicTrackDeviceReader) BuildMusicTrackDeviceMap(musicTrackDevice *MusicTrackDeviceXML) MusicTrackDeviceMap {
	musicTrackDeviceMap := make(MusicTrackDeviceMap)

	for _, entry := range musicTrackDevice.Attributes.ChannelGroup.MusicTrackChannel {
		if entry.Connection == nil {
			continue
		}

		var objectID string
		for _, connectionEntry := range entry.Connection {
			if connectionEntry.ObjectID == "" {
				continue
			}

			if connectionEntry.XID == "instrumentOut" {
				objectID = connectionEntry.ObjectID
			}
		}
		if objectID == "" {
			slog.Error("Object ID is empty")
			continue
		}

		pattern := `(.*)\/Input`

		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(objectID)

		var musicTrackDeviceId string
		if len(matches) > 1 {
			musicTrackDeviceId = matches[1]
		} else {
			slog.Error("No regex matches found for object ID")
			continue
		}

		var songID string
		for _, UIDEntry := range entry.UID {
			if UIDEntry.XID == "uniqueID" {
				songID = UIDEntry.UID
			}
		}
		if songID == "" {
			slog.Error("Song ID is empty")
			continue
		}

		musicTrackDeviceMap[musicTrackDeviceId] = &MusicTrackDeviceMapEntry{
			MusicTrackDeviceID: musicTrackDeviceId,
			SongID:             songID,
		}
	}

	return musicTrackDeviceMap
}
