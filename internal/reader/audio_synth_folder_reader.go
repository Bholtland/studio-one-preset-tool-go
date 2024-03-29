package reader

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	"bholtland/studio-one-preset-tool-go/internal/file"
	"encoding/xml"
	"log/slog"
	"path"
	"regexp"
)

type AudioSynthFolderXML struct {
	XMLName    xml.Name `xml:"AudioSynthFolder"`
	Attributes []struct {
		Attributes []struct {
			XID  string `xml:"id,attr"`
			Name string `xml:"name,attr"`
			UID  []struct {
				XID string `xml:"id,attr"`
				UID string `xml:"uid,attr"`
			} `xml:"UID"`
			Attributes []struct {
				XID         string `xml:"id,attr"`
				Name        string `xml:"name,attr"`
				Category    string `xml:"category,attr"`
				SubCategory string `xml:"subCategory,attr"`
			} `xml:"Attributes"`
		} `xml:"Attributes"`
		UID []struct {
			XID string `xml:"id,attr"`
			UID string `xml:"uid,attr"`
		} `xml:"UID"`
		List []struct {
			XID string `xml:"id,attr"`
			UID struct {
				UID string `xml:"uid,attr"`
			} `xml:"UID"`
		} `xml:"List"`
		String []struct {
			XID  string `xml:"id,attr"`
			Text string `xml:"text,attr"`
		} `xml:"String"`
	} `xml:"Attributes"`
}

type AudioSynthFolderMap map[string]*AudioSynthFolderMapEntry

type AudioSynthFolderMapEntry struct {
	MusicTrackDeviceID string
	DeviceClassID      string
	DeviceName         string
	DeviceUID          string
	DeviceCategory     string
	DeviceSubCategory  string
	DeviceBaseName     string
	PresetPath         string
	PresetFileName     string
}

type AudioSynthFolderReader struct {
	cfg *config.Config
}

func NewAudioSynthFolderReader(cfg *config.Config) *AudioSynthFolderReader {
	return &AudioSynthFolderReader{
		cfg: cfg,
	}
}

func (s *AudioSynthFolderReader) GetMap() (AudioSynthFolderMap, error) {
	xml, err := file.ReadXML[AudioSynthFolderXML](path.Join(s.cfg.Temp.SongContentsPath, "Devices", "audiosynthfolder.xml"))
	if err != nil {
		return nil, err
	}

	return s.buildAudioSynthFolderMap(xml), nil
}

func (s *AudioSynthFolderReader) buildAudioSynthFolderMap(audioSynthFolder *AudioSynthFolderXML) AudioSynthFolderMap {
	audioSynthFolderMap := make(AudioSynthFolderMap)

	for _, entry := range audioSynthFolder.Attributes {
		var musicTrackDeviceID string
		for _, tag := range entry.List {
			if tag.XID == "synthChannels" {
				musicTrackDeviceID = tag.UID.UID
			}
		}
		if musicTrackDeviceID == "" {
			slog.Error("Music Track Device ID is empty")
			continue
		}

		var deviceClassID string
		for _, tag := range entry.UID {
			if tag.XID == "deviceClassID" {
				deviceClassID = tag.UID
			}
		}
		if deviceClassID == "" {
			slog.Error("Device Class ID is empty")
			continue
		}

		var deviceName string
		var deviceUID string
		var deviceCategory string
		var deviceSubCategory string
		var deviceBaseName string
		for _, tag := range entry.Attributes {
			if tag.XID == "deviceData" {
				deviceName = tag.Name

				for _, uidTag := range tag.UID {
					if uidTag.XID == "uniqueID" {
						deviceUID = uidTag.UID
					}
				}
			}
			if tag.XID == "ghostData" {
				for _, attrTag := range tag.Attributes {
					if attrTag.XID == "classInfo" {
						deviceCategory = attrTag.Category
						deviceSubCategory = attrTag.SubCategory
						deviceBaseName = attrTag.Name
					}
				}
			}
		}
		if deviceName == "" {
			slog.Error("Device Name is empty")
			continue
		}
		if deviceUID == "" {
			slog.Error("Device UID is empty")
			continue
		}
		if deviceCategory == "" {
			slog.Error("Device Category is empty")
			continue
		}
		if deviceSubCategory == "" {
			slog.Error("Device Sub Category is empty")
			continue
		}
		if deviceBaseName == "" {
			slog.Error("Device Base Name is empty")
			continue
		}

		var presetPath string
		for _, tag := range entry.String {
			if tag.XID == "presetPath" {
				presetPath = tag.Text
			}
		}
		if presetPath == "" {
			slog.Error("Preset Path is empty")
			continue
		}

		pattern := `.*/([^/]+)$`
		regex := regexp.MustCompile(pattern)
		matches := regex.FindStringSubmatch(presetPath)

		var presetFileName string
		if len(matches) > 1 {
			presetFileName = matches[1]
		} else {
			slog.Error("No regex matches found for preset path")
			continue
		}

		audioSynthFolderMap[musicTrackDeviceID] = &AudioSynthFolderMapEntry{
			MusicTrackDeviceID: musicTrackDeviceID,
			DeviceClassID:      deviceClassID,
			DeviceName:         deviceName,
			DeviceUID:          deviceUID,
			DeviceCategory:     deviceCategory,
			DeviceSubCategory:  deviceSubCategory,
			DeviceBaseName:     deviceBaseName,
			PresetPath:         presetPath,
			PresetFileName:     presetFileName,
		}
	}

	return audioSynthFolderMap
}
