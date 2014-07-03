package mpdinfo

import (
	"fmt"
)

type CurrentSong struct {
	Artist, AlbumArtist, ArtistSort, AlbumArtistSort, Title, Album, Track, Genre string
	Position, Id                                                                 uint
}

func (this *CurrentSong) String() string {
	return fmt.Sprintf("%s by %s (from %s)", this.Artist, this.Title, this.Album)
}
