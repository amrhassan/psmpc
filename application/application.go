package application

import (
	"fmt"
	"github.com/amrhassan/psmpc/gui"
	"github.com/amrhassan/psmpc/mpd"
	"time"
)

type Application struct {
	gui           *gui.GUI
	player        *mpd.Player
	quitRequested bool
}

func NewApplication() *Application {
	return &Application{
		gui:           gui.NewGUI(),
		player:        mpd.NewPlayer(),
		quitRequested: false,
	}
}

func (this *Application) Run() {
	this.player.Connect()
	defer this.player.Disconnect()

	this.gui.RegisterActionHandler(gui.ACTION_PLAYPAUSE, func(args []interface{}) {
		this.player.PlayPause()
	})

	this.gui.RegisterActionHandler(gui.ACTION_NEXT, func(args []interface{}) {
		this.player.Next()
	})

	this.gui.RegisterActionHandler(gui.ACTION_PREVIOUS, func(args []interface{}) {
		this.player.Previous()
	})

	this.gui.RegisterActionHandler(gui.ACTION_QUIT, func(args []interface{}) {
		this.quitRequested = true
	})

	this.updateGui()

	this.player.RegisterChangeHandler(func() {
		this.updateGui()
	})

	go this.gui.Run()
	defer this.gui.Quit()

	for !this.quitRequested {
		time.Sleep(1 * time.Second)
	}

}

func (this *Application) updateGui() {
	current_song, _ := this.player.GetCurrentSong()
	status, _ := this.player.GetStatus()
	this.gui.UpdateCurrentSong(current_song)
	this.gui.UpdateCurrentStatus(status)
}
