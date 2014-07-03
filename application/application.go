package application

import (
	"github.com/amrhassan/psmpc/gui"
	"github.com/amrhassan/psmpc/logging"
	"github.com/amrhassan/psmpc/mpd"
	"github.com/amrhassan/psmpc/mpdinfo"
	"time"
)

var logger = logging.New("app")

type Application struct {
	gui           *gui.GUI
	player        *mpd.Player
	quitRequested bool
}

func NewApplication() *Application {

	return &Application{
		player:        mpd.NewPlayer(),
		quitRequested: false,
	}
}

func get_keymap() map[int]gui.Action {
	return map[int]gui.Action{
		80: gui.ACTION_PLAYPAUSE, // P
		60: gui.ACTION_PREVIOUS,  // <
		62: gui.ACTION_NEXT,      // >
	}
}

func (this *Application) runGui() {

	this.gui = gui.NewGUI(get_keymap())
	defer this.gui.Quit()

	this.gui.RegisterActionHandler(gui.ACTION_PLAYPAUSE, func(args []interface{}) {
		err := this.player.PlayPause()
		if err != nil {
			logger.Fatal("Failed to playpause: %v", err)
		}
	})

	this.gui.RegisterActionHandler(gui.ACTION_NEXT, func(args []interface{}) {
		err := this.player.Next()
		if err != nil {
			logger.Fatal("Failed to next: %v", err)
		}
	})

	this.gui.RegisterActionHandler(gui.ACTION_PREVIOUS, func(args []interface{}) {
		err := this.player.Previous()
		if err != nil {
			logger.Fatal("Failed to previous: %v", err)
		}
	})

	this.gui.RegisterActionHandler(gui.ACTION_QUIT, func(args []interface{}) {
		this.quitRequested = true
	})

	this.updateGui()

	this.player.RegisterChangeHandler(func() {
		this.updateGui()
	})

	go func() {
		status, err := this.player.GetStatus()
		if err != nil {
			logger.Fatal("Failed to connect to MPD: %s", err)
		}

		for {
			if status.State == mpdinfo.STATE_PLAYING {
				this.gui.UpdateCurrentStatus(status)
			}
			time.Sleep(3 * time.Second)
			status, err = this.player.GetStatus()
			if err != nil {
				logger.Fatal("Failed to connect to MPD: %s", err)
			}
		}
	}()

	this.gui.Run() // This blocks the goroutine
}

func (this *Application) Run() {
	this.player.Connect()

	go this.runGui()

	for !this.quitRequested {
		time.Sleep(1 * time.Second)
	}

	logger.Info("Bye")
}

func (this *Application) updateGui() {

	logger.Debug("About to update the GUI")

	current_song, _ := this.player.GetCurrentSong()
	status, _ := this.player.GetStatus()

	if status != nil {
		this.gui.UpdateCurrentStatus(status)
	}

	if status.State != mpdinfo.STATE_STOPPED && current_song != nil {
		this.gui.UpdateCurrentSong(current_song)
	}
}
