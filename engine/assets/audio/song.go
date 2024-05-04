package audio

import (
	"log"

	"github.com/ebitengine/oto/v3"
	"github.com/jfreymuth/oggvorbis"
	"tophatdemon.com/total-invasion-ii/engine/assets"
)

type Song struct {
	oto.Player
	songReader SongReader
}

func LoadSong(assetPath string) (*Song, error) {
	file, err := assets.GetFile(assetPath)
	if err != nil {
		return nil, err
	}
	//defer file.Close()

	oggReader, err := oggvorbis.NewReader(file)
	if err != nil {
		return nil, err
	}

	song := &Song{
		songReader: NewSongReader(*oggReader, true),
	}
	song.Player = *context.NewPlayer(&song.songReader)

	log.Printf("Song loaded at %v.\n", assetPath)
	return song, nil
}

func (song *Song) Free() {

}
