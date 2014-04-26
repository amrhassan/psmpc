package gui

import (
	"container/list"
	"github.com/amrhassan/psmpc/mpdinfo"
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
)

type GUI struct {
	builder                    *gtk.Builder
	main_window                *gtk.Window
	title_label                *gtk.Label
	artist_label               *gtk.Label
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

	return &GUI{
		builder:                    builder,
		main_window:                main_window,
		title_label:                title_label,
		artist_label:               artist_label,
		registered_action_handlers: make(map[Action]*list.List),
	}
}

// Initiates the GUI
func (this *GUI) Run() {

	this.main_window.ShowAll()

	this.main_window.Connect("destroy", func() {
		gtk.MainQuit()
	})

	playpause_button, err := this.builder.GetObject("play-pause_button")
	if err != nil {
		error_panic("Failed to retrieve play-pause_button", err)
	}

	playpause_button.(*gtk.Button).Connect("clicked", func() {
		this.fireAction(ACTION_PLAYPAUSE)
	})

	gtk.Main()
}

// Shuts down the GUI
func (this *GUI) Quit() {
	gtk.MainQuit()
}

// Updates the GUI with the currently-playing song information
func (this *GUI) UpdateCurrentSong(current_song *mpdinfo.CurrentSong) {

	if current_song.Title != "" {
		this.title_label.SetText(current_song.Title)
	}

	if current_song.Artist != "" {
		this.artist_label.SetText(current_song.Artist)
	}
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
