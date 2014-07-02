package resources

import (
	"github.com/amrhassan/psmpc/mpdinfo"
	"github.com/rakyll/magicmime"
	"io/ioutil"
	"strings"
	"testing"
)

func TestGetResourceWithArtistAndTitle(t *testing.T) {
	track := &Track{
		&mpdinfo.CurrentSong{
			Artist: "Radiohead",
			Title:  "Paranoid Android",
		},
	}

	testGetResource(track, t)
}

func TestGetResourceWithAlbumAndTitle(t *testing.T) {
	track := &Track{
		&mpdinfo.CurrentSong{
			Artist: "Radiohead",
			Album:  "OK Computer",
		},
	}

	testGetResource(track, t)
}

func testGetResource(track *Track, t *testing.T) {
	provider := newLastFMAlbumArtProvider()

	resource, err := provider.GetResource(track)

	if err != nil {
		t.Error(err)
	}

	content, err := ioutil.ReadAll(resource)

	if err != nil {
		t.Error(err)
	}

	magic, err := magicmime.New()
	if err != nil {
		t.Error(err)
	}

	magicType, _ := magic.TypeByBuffer(content)

	if !strings.HasPrefix(magicType, "image") {
		t.Errorf("The returned resource stream has this type: %s", magicType)
	}
}

func TestGetType(t *testing.T) {
	provider := newLastFMAlbumArtProvider()
	if provider.GetType() != ALBUM_ART {
		t.Error("Provider is not ALBUM_ART")
	}
}
