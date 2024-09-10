package music

import (
	"bytes"
	"log"
	"mime"
	"os"
	"path/filepath"

	"github.com/dhowden/tag"
)

type MusicMetaData struct {
	filepath    string
	fileformat  string
	title       string
	artist      string
	album       string
	albumartist string
	composer    string
	year        int
	genre       string
	lyrics      string
	comment     string
	track       struct {
		number int
		total  int
	}
	discs struct {
		number int
		total  int
	}
}

func GetFileType(filename string) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}

	// Check magic bytes
	if bytes.Equal(data[:4], []byte("RIFF")) {
		return "audio/wav", nil
	} else if bytes.Equal(data[:3], []byte("ID3")) {
		return "audio/mp3", nil
	} else if bytes.Equal(data[:4], []byte("FLaC")) {
		return "audio/flac", nil
	} else if bytes.Equal(data[:4], []byte("Oggs")) {
		return "audio/ogg", nil
	}
	// Fallback to MIME type based on extension
	ext := filepath.Ext(filename)
	mimeType := mime.TypeByExtension(ext)
	if err != nil {
		return "", err
	}

	return mimeType, nil
}

func GetMetadata(filepath string) *MusicMetaData {
	music := MusicMetaData{}
	osfile, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	music.fileformat, err = GetFileType(filepath)
	if err != nil {
		log.Fatal(err)
	}
	// early return as tag doesnt support wav fileformats
	if music.fileformat == "audio/wav" {
		return &music
	}
	m, err := tag.ReadFrom(osfile)
	if err != nil {
		return &music
	}
	music.filepath = filepath
	music.title = m.Title()
	music.artist = m.Artist()
	music.album, music.albumartist = m.Album(), m.AlbumArtist()
	music.composer = m.Composer()
	music.year = m.Year()
	music.genre = m.Genre()
	music.lyrics, music.comment = m.Lyrics(), m.Comment()
	music.track.number, music.track.total = m.Track()
	music.discs.number, music.discs.total = m.Disc()
	return &music
}
