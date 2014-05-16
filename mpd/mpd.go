package mpd

import (
	"code.google.com/p/gompd/mpd"
	"github.com/amrhassan/psmpc/mpdinfo"
)

type Player struct {
	client *mpd.Client
}

// Connects to the MPD server located at localhost:6600. Returns a Player instance.
func Connect() *Player {
	client, _ := mpd.Dial("tcp", "localhost:6600")
	return &Player{
		client,
	}
}

// Disconnects this connection
func (this *Player) Disconnect() error {
	return this.client.Close()
}

func (this *Player) GetCurrentSong() (*mpdinfo.CurrentSong, error) {
	current_song, err := this.client.CurrentSong()
	if err != nil {
		return nil, err
	}

	return &mpdinfo.CurrentSong{
		Title:  current_song["Title"],
		Artist: current_song["Artist"],
	}, nil
}

func (this *Player) PlayPause() error {
	current_status, err := this.GetStatus()
	if err != nil {
		return nil
	}

	if current_status.State == mpdinfo.STATE_PLAYING {
		return this.client.Pause(true)
	} else {
		return this.client.Pause(false)
	}
}

func (this *Player) GetStatus() (*mpdinfo.Status, error) {
	status, err := this.client.Status()
	if err != nil {
		return nil, err
	}

	return &mpdinfo.Status{
		State: mpdinfo.State(status["state"]),
	}, nil
}

func (this *Player) Next() error {
	return this.client.Next()
}

func (this *Player) Previous() error {
	return this.client.Previous()
}
