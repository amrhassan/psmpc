package gui

import (
	"container/list"
	"github.com/amrhassan/psmpc/mpdinfo"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
)

const glad_file_path = "gui/ui.glade"

// An action type label
type Action int

// Handlers for fired actions from the GUI
type ActionHandler func([]interface{})

// Player actions that can be performed from the GUI
const (
	ACTION_PLAY      = iota
	ACTION_PAUSE     = iota
	ACTION_PLAYPAUSE = iota
	ACTION_NEXT      = iota
	ACTION_PREVIOUS  = iota
	ACTION_QUIT      = iota
)

type GUI struct {
	builder                    *gtk.Builder
	main_window                *gtk.Window
	title_label                *gtk.Label
	artist_label               *gtk.Label
	stopped_header             *gtk.Box
	playback_header            *gtk.Box
	registered_action_handlers map[Action]*list.List
}

func error_panic(message string, err error) {
	panic(message + ": " + err.Error())
}

// Constructs and initializes a new GUI instance
func NewGUI() *GUI {
	gtk.Init(nil)

	builder, err := gtk.BuilderNew()
	if err != nil {
		error_panic("Failed to create gtk.Builder", err)
	}

	err = builder.AddFromFile(glad_file_path)
	if err != nil {
		error_panic("Failed to load the Glade UI file", err)
	}

	main_window_gobject, err := builder.GetObject("main_window")
	if err != nil {
		error_panic("Failed to retrieve the main_window object", err)
	}

	main_window := main_window_gobject.(*gtk.Window)

	artist_label_gobject, err := builder.GetObject("artist_label")
	if err != nil {
		error_panic("Failed to retrieve artist_label object", err)
	}
	artist_label := artist_label_gobject.(*gtk.Label)

	title_label_gobject, err := builder.GetObject("title_label")
	if err != nil {
		error_panic("Failed to retrieve title_label object", err)
	}
	title_label := title_label_gobject.(*gtk.Label)

	stopped_header_gobject, err := builder.GetObject("stopped_header_box")
	if err != nil {
		error_panic("Failed to retrieve stopped_header_box gobject", err)
	}
	stopped_header := stopped_header_gobject.(*gtk.Box)

	playback_header_gobject, err := builder.GetObject("playback_header_box")
	if err != nil {
		error_panic("Failed to retrieve playback_header_box", err)
	}
	playback_header := playback_header_gobject.(*gtk.Box)

	return &GUI{
		builder:                    builder,
		main_window:                main_window,
		title_label:                title_label,
		artist_label:               artist_label,
		stopped_header:             stopped_header,
		playback_header:            playback_header,
		registered_action_handlers: make(map[Action]*list.List),
	}
}

func (this *GUI) getGtkObject(name string) glib.IObject {
	object, err := this.builder.GetObject(name)
	if err != nil {
		panic("Failed to retrieve GTK object " + name)
	}
	return object
}

// Initiates the GUI
func (this *GUI) Run() {

	this.main_window.Connect("destroy", func() {
		this.fireAction(ACTION_QUIT)
	})

	this.getGtkObject("play-pause_button").(*gtk.Button).Connect("clicked", func() {
		this.fireAction(ACTION_PLAYPAUSE)
	})

	this.getGtkObject("previous_button").(*gtk.Button).Connect("clicked", func() {
		this.fireAction(ACTION_PREVIOUS)
	})

	this.getGtkObject("next_button").(*gtk.Button).Connect("clicked", func() {
		this.fireAction(ACTION_NEXT)
	})

	this.main_window.ShowAll()
	gtk.Main()

}

// Shuts down the GUI
func (this *GUI) Quit() {
	glib.IdleAdd(func() {
		gtk.MainQuit()
	})
}

// Updates the GUI with the currently-playing song information
func (this *GUI) UpdateCurrentSong(current_song *mpdinfo.CurrentSong) {
	glib.IdleAdd(func() {
		if current_song.Title != "" {
			this.title_label.SetText(current_song.Title)
		}

		if current_song.Artist != "" {
			this.artist_label.SetText(current_song.Artist)
		}
	})
}

// Updates the GUI with the current MPD status
func (this *GUI) UpdateCurrentStatus(current_status *mpdinfo.Status) {
	glib.IdleAdd(func() {
		switch current_status.State {

		case mpdinfo.STATE_STOPPED:
			this.stopped_header.Show()
			this.playback_header.Hide()

		case mpdinfo.STATE_PLAYING:
			pause_image := this.getGtkObject("pause_image").(*gtk.Image)
			this.getGtkObject("play-pause_button").(*gtk.Button).SetImage(pause_image)
			this.stopped_header.Hide()
			this.playback_header.Show()

		case mpdinfo.STATE_PAUSED:
			play_image := this.getGtkObject("play_image").(*gtk.Image)
			this.getGtkObject("play-pause_button").(*gtk.Button).SetImage(play_image)
			this.stopped_header.Hide()
			this.playback_header.Show()
		}
	})
}

// Fires the action specified by the given Action, passing the given arguments to all the
// subscribed handlers
func (this *GUI) fireAction(action_type Action, args ...interface{}) {
	handlers, any := this.registered_action_handlers[action_type]

	if any == false {
		// None are registered
		return
	}

	for e := handlers.Front(); e != nil; e = e.Next() {
		handler := e.Value.(ActionHandler)
		handler(args)
	}
}

// Registers an action handler to be executed when a specific action is fired
func (this *GUI) RegisterActionHandler(action_type Action, handler ActionHandler) {
	_, handlers_exist := this.registered_action_handlers[action_type]
	if !handlers_exist {
		this.registered_action_handlers[action_type] = list.New()
	}

	this.registered_action_handlers[action_type].PushFront(handler)
}
