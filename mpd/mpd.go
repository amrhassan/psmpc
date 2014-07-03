package mpd

import (
	"code.google.com/p/gompd/mpd"
	"container/list"
	"errors"
	"github.com/amrhassan/psmpc/logging"
	"github.com/amrhassan/psmpc/mpdinfo"
	"strconv"
	"strings"
)

var logger = logging.New("mpd")

type Player struct {
	hostname       string
	port           uint
	changeHandlers *list.List
}

type ChangeHandler func()

var playerNotConnectedError = errors.New("This player is not connected")

// Returns a Player instance to a server at localhost:6600.
func NewPlayer() *Player {
	// TODO: Read hostname and port from environment vars MPD_HOST and MPD_PORT
	hostname := "localhost"
	port := uint(6600)

	return &Player{
		hostname:       hostname,
		port:           port,
		changeHandlers: list.New(),
	}
}

func (this *Player) establishConnection() (*mpd.Client, error) {
	client, err := mpd.Dial("tcp", "localhost:6600")
	if err != nil {
		return nil, err
	}
	return client, nil
}

// Connects to the MPD server.
func (this *Player) Connect() error {

	watcher, err := mpd.NewWatcher("tcp", "localhost:6600", "")
	if err != nil {
		return err
	}

	go func() {
		for event := range watcher.Event {
			logger.Debug("Got MPD event:", event)
			for e := this.changeHandlers.Front(); e != nil; e = e.Next() {
				logger.Debug("Notifying change handler:", e.Value)
				e.Value.(ChangeHandler)()
			}
		}
	}()

	return nil
}

/*
 * Returns nil if no current song is playing
 */
func (this *Player) GetCurrentSong() (*mpdinfo.CurrentSong, error) {

	client, err := this.establishConnection()
	defer client.Close()
	if err != nil {
		return nil, err
	}

	current_song, err := client.CurrentSong()
	if err != nil {
		return nil, err
	}

	if current_song["Title"] == "" { // Unacceptable
		return nil, nil
	}

	return &mpdinfo.CurrentSong{
		Title:  current_song["Title"],
		Artist: current_song["Artist"],
		Album:  current_song["Album"],
	}, nil
}

func (this *Player) PlayPause() error {

	client, err := this.establishConnection()
	defer client.Close()
	if err != nil {
		return err
	}

	current_status, err := this.GetStatus()
	if err != nil {
		return err
	}

	if current_status.State == mpdinfo.STATE_PLAYING {
		return client.Pause(true)
	} else {
		return client.Pause(false)
	}
}

func (this *Player) GetStatus() (*mpdinfo.Status, error) {

	client, err := this.establishConnection()
	defer client.Close()
	if err != nil {
		return nil, err
	}

	status, err := client.Status()
	if err != nil {
		return nil, err
	}

	time := strings.Split(status["time"], ":")
	elapsed, err := strconv.ParseFloat(time[0], 32)
	if err != nil {
		logger.Warn("Failed to parse the value of the elapsed song time: %s", time[0])
		elapsed = 0.0
	}
	totalTime, err := strconv.ParseFloat(time[1], 32)
	if err != nil {
		logger.Warn("Failed to parse the value of the total song length: %s", time[1])
		totalTime = 1.0
	}

	return &mpdinfo.Status{
		State:        mpdinfo.State(status["state"]),
		SongProgress: elapsed / totalTime,
	}, nil
}

func (this *Player) Next() error {
	client, err := this.establishConnection()
	defer client.Close()
	if err != nil {
		return err
	}

	return client.Next()
}

func (this *Player) Previous() error {
	client, err := this.establishConnection()
	defer client.Close()
	if err != nil {
		return err
	}

	return client.Previous()
}

func (this *Player) RegisterChangeHandler(changeHandler ChangeHandler) {
	this.changeHandlers.PushFront(changeHandler)
}
