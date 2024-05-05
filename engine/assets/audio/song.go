package audio

import (
	"log"
	"os"

	"github.com/ebitengine/oto/v3"
	"github.com/jfreymuth/oggvorbis"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type Song struct {
	oto.Player
	songReader SongReader
	file       *os.File
}

func LoadSong(assetPath string) (*Song, error) {
	file, err := assets.GetFile(assetPath)
	if err != nil {
		return nil, err
	}

	oggReader, err := oggvorbis.NewReader(file)
	if err != nil {
		return nil, err
	}

	song := &Song{
		file:       file,
		songReader: NewSongReader(*oggReader, true),
	}
	song.Player = *context.NewPlayer(&song.songReader)

	log.Printf("Song loaded at %v.\n", assetPath)
	return song, nil
}

func (song *Song) Free() {
	song.file.Close()
}
