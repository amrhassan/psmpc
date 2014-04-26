package main

import (
	"github.com/amrhassan/psmpc/gui"
	"github.com/amrhassan/psmpc/mpd"
	"time"
)

func main() {
	var player = mpd.Connect()
	defer player.Disconnect()

	var g = gui.NewGUI()
	go g.Run()

	g.RegisterActionHandler(gui.ACTION_PLAYPAUSE, func(args []interface{}) {
		player.PlayPause()
	})

	for {
		current_song, _ := player.GetCurrentSong()
		g.UpdateCurrentSong(current_song)
		time.Sleep(1 * time.Second)
	}
}
