package application

import (
	"github.com/amrhassan/psmpc/gui"
	"github.com/amrhassan/psmpc/mpd"
	"log"
	"time"
)

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
			log.Fatalf("Failed to playpause: %v", err)
		}
	})

	this.gui.RegisterActionHandler(gui.ACTION_NEXT, func(args []interface{}) {
		err := this.player.Next()
		if err != nil {
			log.Fatalf("Failed to next: %v", err)
		}
	})

	this.gui.RegisterActionHandler(gui.ACTION_PREVIOUS, func(args []interface{}) {
		err := this.player.Previous()
		if err != nil {
			log.Fatalf("Failed to previous: %v", err)
		}
	})

	this.gui.RegisterActionHandler(gui.ACTION_QUIT, func(args []interface{}) {
		this.quitRequested = true
	})

	this.updateGui()

	this.player.RegisterChangeHandler(func() {
		this.updateGui()
	})

	this.gui.Run()
}

func (this *Application) Run() {
	this.player.Connect()

	go this.runGui()

	for !this.quitRequested {
		time.Sleep(1 * time.Second)
	}

	log.Println("Bye")
}

func (this *Application) updateGui() {

	log.Printf("About to update the GUI")

	current_song, _ := this.player.GetCurrentSong()
	status, _ := this.player.GetStatus()

	if current_song != nil {
		this.gui.UpdateCurrentSong(current_song)
	}

	if status != nil {
		this.gui.UpdateCurrentStatus(status)
	}
}
