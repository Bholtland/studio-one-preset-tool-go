package reader

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	"bholtland/studio-one-preset-tool-go/internal/file"
	"encoding/xml"
	"log/slog"
	"path"
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
	cfg *config.Config
}

func NewSongReader(cfg *config.Config) *SongReader {
	return &SongReader{
		cfg: cfg,
	}
}

func (s *SongReader) GetMap() (SongMap, FolderMap, error) {
	xml, err := file.ReadXML[SongXML](path.Join(s.cfg.Temp.SongContentsPath, "Song", "song.xml"))
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
