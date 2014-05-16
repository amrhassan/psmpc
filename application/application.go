package application

import (
	"github.com/amrhassan/psmpc/gui"
	"github.com/amrhassan/psmpc/mpd"
	"time"
)

type Application struct {
	gui           *gui.GUI
	quitRequested bool
}

func NewApplication() *Application {
	return &Application{
		gui:           gui.NewGUI(),
		quitRequested: false,
	}
}

func (this *Application) Run() {
	player := mpd.NewPlayer()
	player.Connect()
	defer player.Disconnect()

	go this.gui.Run()
	defer this.gui.Quit()

	this.gui.RegisterActionHandler(gui.ACTION_PLAYPAUSE, func(args []interface{}) {
		player.PlayPause()
	})

	this.gui.RegisterActionHandler(gui.ACTION_NEXT, func(args []interface{}) {
		player.Next()
	})

	this.gui.RegisterActionHandler(gui.ACTION_PREVIOUS, func(args []interface{}) {
		player.Previous()
	})

	this.gui.RegisterActionHandler(gui.ACTION_QUIT, func(args []interface{}) {
		this.quitRequested = true
	})

	for !this.quitRequested {
		current_song, _ := player.GetCurrentSong()
		status, _ := player.GetStatus()
		this.gui.UpdateCurrentSong(current_song)
		this.gui.UpdateCurrentStatus(status)
		time.Sleep(1 * time.Second)
	}

}
