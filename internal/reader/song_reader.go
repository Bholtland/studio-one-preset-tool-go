package reader

import (
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type SongXML struct {
	XMLName    xml.Name `xml:"Song"`
	Attributes struct {
		List struct {
			XID         string `xml:"id,attr"`
			MediaTracks []struct {
				TrackID      string `xml:"trackID,attr"`
				ParentFolder string `xml:"parentFolder,attr"`
				Name         string `xml:"name,attr"`
				UID          []struct {
					XID string `xml:"id,attr"`
					UID string `xml:"uid,attr"`
				}
			} `xml:"MediaTrack"`
			FolderTracks []struct {
				ParentTrackID string `xml:"parentFolder,attr"`
				TrackID       string `xml:"trackID,attr"`
				Name          string `xml:"name,attr"`
			} `xml:"FolderTrack"`
		} `xml:"List"`
	} `xml:"Attributes"`
}

type SongMap map[string]*SongMapEntry

type SongMapEntry struct {
	ParentTrackID string
	Name          string
	TrackID       string
}

type FolderMap map[string]*FolderMapEntry

type FolderMapEntry struct {
	Name          string
	ParentTrackID string
}

type SongReader struct {
}

func NewSongReader() *SongReader {
	return &SongReader{}
}

func (s *SongReader) GetMap() (SongMap, FolderMap, error) {
	xml, err := s.getXML()
	if err != nil {
		return nil, nil, err
	}

	return s.buildSongMap(xml), s.buildFolderMap(xml), nil
}

func (s *SongReader) buildSongMap(song *SongXML) SongMap {
	songMap := make(map[string]*SongMapEntry)

	for _, entry := range song.Attributes.List.MediaTracks {
		var songID string
		for _, uidEntry := range entry.UID {
			if uidEntry.XID == "channelID" {
				songID = uidEntry.UID
			}
		}

		if songID == "" {
			slog.Error("uid is empty")
			continue
		}

		// TODO: Handle not founds

		songMap[songID] = &SongMapEntry{
			TrackID:       entry.TrackID,
			Name:          entry.Name,
			ParentTrackID: entry.ParentFolder,
		}
	}

	return songMap
}

func (s *SongReader) buildFolderMap(song *SongXML) map[string]*FolderMapEntry {
	var presetsMap = make(map[string]*FolderMapEntry)

	for _, track := range song.Attributes.List.FolderTracks {
		if track.Name == "" {
			slog.Error("Track name is empty")
			continue
		}
		if track.TrackID == "" {
			slog.Error("Track ID is empty")
			continue
		}
		presetsMap[track.TrackID] = &FolderMapEntry{
			Name:          track.Name,
			ParentTrackID: track.ParentTrackID,
		}
	}
	return presetsMap
}

func (sr *SongReader) getXML() (*SongXML, error) {
	file, err := os.Open("song.xml")
	if err != nil {
		return nil, fmt.Errorf("Error opening XML file: %w", err)
	}
	defer file.Close()

	// Read the entire XML content
	rawXML, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading XML file: %w", err)
	}

	var unmarshalledXML *SongXML

	// Unmarshal the XML content into the Person struct
	err = xml.Unmarshal(rawXML, &unmarshalledXML)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling XML: %w", err)
	}

	return unmarshalledXML, nil
}