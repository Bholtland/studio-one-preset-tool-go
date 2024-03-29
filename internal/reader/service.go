package reader

import (
	"bholtland/studio-one-preset-tool-go/internal/config"
	"log/slog"
)

type PresetMap map[string]*PresetMapEntry

type PresetMapEntry struct {
	DeviceClassID     string
	DeviceBaseName    string
	DeviceCategory    string
	DeviceSubCategory string
	DeviceName        string
	DeviceUID         string
	TrackID           string
	FileName          string
	Name              string
	Path              string
	SongID            string
}

type Service struct {
	audioSynthFolderReader *AudioSynthFolderReader
	musicTrackDeviceReader *MusicTrackDeviceReader
	songReader             *SongReader
	cfg                    *config.Config
}

func NewService(cfg *config.Config) *Service {
	return &Service{
		audioSynthFolderReader: NewAudioSynthFolderReader(cfg),
		musicTrackDeviceReader: NewMusicTrackDeviceReader(cfg),
		songReader:             NewSongReader(cfg),
		cfg:                    cfg,
	}
}

func (s *Service) GetPresets() (PresetMap, error) {
	songMap, folderMap, err := s.songReader.GetMap()
	if err != nil {
		return nil, err
	}

	audioSynthFolderMap, err := s.audioSynthFolderReader.GetMap()
	if err != nil {
		return nil, err
	}

	musicTrackDeviceMap, err := s.musicTrackDeviceReader.GetMap()
	if err != nil {
		return nil, err
	}

	var presetMap = make(PresetMap)

	for _, audioSynthFolderEntry := range audioSynthFolderMap {
		musicTrackDeviceEntry, ok := musicTrackDeviceMap[audioSynthFolderEntry.MusicTrackDeviceID]
		if !ok {
			slog.Error("Music Track Device not found for track")
			continue
		}

		songEntry, ok := songMap[musicTrackDeviceEntry.SongID]
		if !ok {
			slog.Error("Music Track Device not found for track")
			continue
		}

		path := GetPath(songEntry.ParentTrackID, folderMap)

		preset := &PresetMapEntry{
			DeviceClassID:     audioSynthFolderEntry.DeviceClassID,
			DeviceBaseName:    audioSynthFolderEntry.DeviceBaseName,
			DeviceCategory:    audioSynthFolderEntry.DeviceCategory,
			DeviceSubCategory: audioSynthFolderEntry.DeviceSubCategory,
			DeviceName:        audioSynthFolderEntry.DeviceName,
			DeviceUID:         audioSynthFolderEntry.DeviceUID,
			TrackID:           songEntry.TrackID,
			FileName:          audioSynthFolderEntry.PresetFileName,
			Name:              songEntry.Name,
			Path:              path,
			SongID:            musicTrackDeviceEntry.SongID,
		}
		presetMap[audioSynthFolderEntry.MusicTrackDeviceID] = preset
	}

	return presetMap, nil

}

func GetPath(parentTrackID string, folderMap FolderMap) string {
	if parentTrackID == "" {
		return ""
	}

	folder, ok := folderMap[parentTrackID]
	if !ok {
		slog.Error("Folder not found for track")
		return ""
	}

	if folder.ParentTrackID == "" {
		return folder.Name
	}

	return GetPath(folder.ParentTrackID, folderMap) + "/" + folder.Name
}
