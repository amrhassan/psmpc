/*
An ALBUM_ART resource provider from Last.fm
*/

package resources

import (
	"github.com/shkh/lastfm-go/lastfm"
	"io"
	"net/http"
)

const apiKey = "43ffca14ea943af9f30bd147cd03e891"

type LastFMAlbumArtProvider struct {
	api *lastfm.Api
}

func newLastFMAlbumArtProvider() *LastFMAlbumArtProvider {
	return &LastFMAlbumArtProvider{
		api: lastfm.New(apiKey, ""),
	}
}

func (this *LastFMAlbumArtProvider) Type() ResourceType {
	return ALBUM_ART
}

func (this *LastFMAlbumArtProvider) GetResource(track *Track) (io.ReadCloser, error) {

	switch {
	case track.Artist != "" && track.Album != "":
		return this.getAlbumImageUrl(track.Artist, track.Album)
	default:
		return this.getTrackImageUrl(track.Artist, track.Title)
	}
}

func (this *LastFMAlbumArtProvider) getAlbumImageUrl(artist string, album string) (stream io.ReadCloser, err error) {

	args := map[string]interface{}{
		"album":  album,
		"artist": artist,
	}

	albumInfo, err := this.api.Album.GetInfo(args)

	if err != nil {
		return nil, err
	}

	albumImages := albumInfo.Images
	imageUrl := albumImages[len(albumImages)-1].Url

	response, err := http.Get(imageUrl)

	return response.Body, err
}

func (this *LastFMAlbumArtProvider) getTrackImageUrl(artist string, title string) (stream io.ReadCloser, err error) {

	args := map[string]interface{}{
		"track":  title,
		"artist": artist,
	}

	trackInfo, err := this.api.Track.GetInfo(args)

	if err != nil {
		return nil, err
	}

	albumImages := trackInfo.Album.Images
	imageUrl := albumImages[len(albumImages)-1].Url

	response, err := http.Get(imageUrl)

	return response.Body, err
}
