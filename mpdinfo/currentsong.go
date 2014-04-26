package mpdinfo

type CurrentSong struct {
	Artist, AlbumArtist, ArtistSort, AlbumArtistSort, Title, Album, Track, Genre string
	Position, Id                                                                 uint
}
