package mpd

import (
	"code.google.com/p/gompd/mpd"
	"errors"
	"github.com/amrhassan/psmpc/mpdinfo"
)

type Player struct {
	client   *mpd.Client
	hostname string
	port     uint
}

var playerNotConnectedError = errors.New("This player is not connected")

// Returns a Player instance to a server at localhost:6600.
func NewPlayer() *Player {
	// TODO: Read hostname and port from environment vars MPD_HOST and MPD_PORT
	hostname := "localhost"
	port := uint(6600)

	return &Player{
		client:   nil,
		hostname: hostname,
		port:     port,
	}
}

// Connects to the MPD server.
func (this *Player) Connect() error {
	client, error := mpd.Dial("tcp", "localhost:6600")
	this.client = client
	return error
}

// Returns true if this Player is connected to its server
func (this *Player) IsConnected() bool {
	return this.client != nil
}

// Disconnects this connection
func (this *Player) Disconnect() error {
	return this.client.Close()
}

func (this *Player) GetCurrentSong() (*mpdinfo.CurrentSong, error) {

	if !this.IsConnected() {
		return nil, playerNotConnectedError
	}

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

	if !this.IsConnected() {
		return playerNotConnectedError
	}

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

	if !this.IsConnected() {
		return nil, playerNotConnectedError
	}

	status, err := this.client.Status()
	if err != nil {
		return nil, err
	}

	return &mpdinfo.Status{
		State: mpdinfo.State(status["state"]),
	}, nil
}

func (this *Player) Next() error {
	if !this.IsConnected() {
		return playerNotConnectedError
	}

	return this.client.Next()
}

func (this *Player) Previous() error {
	if !this.IsConnected() {
		return playerNotConnectedError
	}

	return this.client.Previous()
}
