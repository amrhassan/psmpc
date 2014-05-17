package gui

/*
#cgo pkg-config: gdk-3.0
#include <gdk/gdk.h>
*/
import "C"

import (
	"container/list"
	"github.com/amrhassan/psmpc/mpdinfo"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
	"log"
	"os"
	"unsafe"
)

/*
 * The paths where the Glade UI file is looked up from. The paths are tried in the order
 * they are listed in, and the first one that exists is used.
 */
var glad_file_paths = []string{
	"gui/ui.glade",
	"~/.local/share/psmpc/gui/ui.glade",
	"/usr/local/share/psmpc/gui/ui.glade",
	"/usr/share/psmpc/gui/ui.glade",
}

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
	buttonKeyMap               map[int]Action
}

func error_panic(message string, err error) {
	panic(message + ": " + err.Error())
}

func get_glade_path() string {

	path_exists := func(path string) bool {
		_, path_error := os.Stat(path)
		return path_error == nil || os.IsExist(path_error)
	}

	for _, path := range glad_file_paths {
		if path_exists(path) {
			log.Println("Using the glade file from: " + path)
			return path
		}
	}

	log.Panic("Can't find a glade UI file")
	return ""
}

// Constructs and initializes a new GUI instance
// Args:
// 	buttonKeyMap: a mapping between button key values and GUI actions
func NewGUI(buttonKeyMap map[int]Action) *GUI {
	gtk.Init(nil)

	builder, err := gtk.BuilderNew()
	if err != nil {
		error_panic("Failed to create gtk.Builder", err)
	}

	err = builder.AddFromFile(get_glade_path())
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
		playback_header:            playback_header,
		registered_action_handlers: make(map[Action]*list.List),
		buttonKeyMap:               buttonKeyMap,
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

	this.main_window.Connect("key-release-event", func(window *gtk.Window, event *gdk.Event) {
		key := extract_key_from_gdk_event(event)
		action, mapped := this.buttonKeyMap[key.value]

		if mapped {
			this.fireAction(action)
		}
	})

	this.main_window.ShowAll()
	gtk.Main()
}

// A keyboard key
type key struct {
	value          int
	representation string
}

// Extracts a key instance from the GdkEventKey wrapped in the given gdk.Event
func extract_key_from_gdk_event(gdk_key_event *gdk.Event) key {
	value := (*C.GdkEventKey)(unsafe.Pointer(gdk_key_event.Native())).keyval
	repr := (*C.char)(C.gdk_keyval_name(value))
	return key{
		value:          int(value),
		representation: C.GoString(repr),
	}
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
			this.getGtkObject("controls_box").(*gtk.Box).Hide()
			this.getGtkObject("artist_box").(*gtk.Box).Hide()
			this.getGtkObject("title_label").(*gtk.Label).SetText("Stopped")

		case mpdinfo.STATE_PLAYING:
			this.getGtkObject("controls_box").(*gtk.Box).Show()
			this.getGtkObject("artist_box").(*gtk.Box).Show()

			pause_image := this.getGtkObject("pause_image").(*gtk.Image)
			this.getGtkObject("play-pause_button").(*gtk.Button).SetImage(pause_image)

		case mpdinfo.STATE_PAUSED:
			this.getGtkObject("controls_box").(*gtk.Box).Show()
			this.getGtkObject("artist_box").(*gtk.Box).Show()

			play_image := this.getGtkObject("play_image").(*gtk.Image)
			this.getGtkObject("play-pause_button").(*gtk.Button).SetImage(play_image)
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
