package gui

/*
#cgo pkg-config: gdk-3.0
#include <gdk/gdk.h>
*/
import "C"

import (
	"container/list"
	"fmt"
	"github.com/amrhassan/psmpc/mpdinfo"
	"github.com/amrhassan/psmpc/resources"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
	"log"
	"os"
	"unsafe"
)

/*
 * The paths where the static resources are looked up from. The paths are tried in the order
 * they are listed in, and the first one that exists is used.
 */
var resource_file_paths = []string{
	".",
	"~/.local/share/psmpc/",
	"/usr/local/share/psmpc/",
	"/usr/share/psmpc/",
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
	controls_box               *gtk.Box
	artist_box                 *gtk.Box
	previous_button            *gtk.Button
	playpause_button           *gtk.Button
	next_button                *gtk.Button
	play_image                 *gtk.Image
	pause_image                *gtk.Image
	status_icon                *gtk.StatusIcon
	album_box                  *gtk.Box
	album_label                *gtk.Label
	album_art_image            *gtk.Image
	registered_action_handlers map[Action]*list.List
	buttonKeyMap               map[int]Action
	resourceManager            *resources.ResourceManager
}

func error_panic(message string, err error) {
	panic(message + ": " + err.Error())
}

func path_exists(path string) bool {
	_, path_error := os.Stat(path)
	return path_error == nil || os.IsExist(path_error)
}

func get_glade_path() string {

	for _, path := range resource_file_paths {
		full_path := path + "/gui/ui.glade"
		if path_exists(full_path) {
			log.Println("Using the glade file from: " + full_path)
			return full_path
		}
	}

	log.Panic("Can't find a glade UI file")
	return ""
}

func get_icon_path() string {
	for _, path := range resource_file_paths {
		full_path := path + "/gui/icon.png"
		if path_exists(full_path) {
			log.Println("Using the icon from: " + full_path)
			return full_path
		}
	}

	log.Panic("Can't find the icon file")
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

	getGtkObject := func(name string) glib.IObject {
		object, err := builder.GetObject(name)
		if err != nil {
			panic("Failed to retrieve GTK object " + name)
		}
		return object
	}

	main_window := getGtkObject("main_window").(*gtk.Window)
	main_window.SetIconFromFile(get_icon_path())

	artist_label := getGtkObject("artist_label").(*gtk.Label)
	title_label := getGtkObject("title_label").(*gtk.Label)
	playback_header := getGtkObject("playback_header_box").(*gtk.Box)
	controls_box := getGtkObject("controls_box").(*gtk.Box)
	artist_box := getGtkObject("artist_box").(*gtk.Box)
	playpause_button := getGtkObject("play-pause_button").(*gtk.Button)
	previous_button := getGtkObject("previous_button").(*gtk.Button)
	next_button := getGtkObject("next_button").(*gtk.Button)
	pause_image, _ := gtk.ImageNewFromIconName("gtk-media-pause", gtk.ICON_SIZE_BUTTON)
	play_image := getGtkObject("play_image").(*gtk.Image)
	status_icon, _ := gtk.StatusIconNewFromFile(get_icon_path())
	album_box := getGtkObject("album_box").(*gtk.Box)
	album_label := getGtkObject("album_label").(*gtk.Label)
	album_art_image := getGtkObject("album_art").(*gtk.Image)

	return &GUI{
		builder:                    builder,
		main_window:                main_window,
		title_label:                title_label,
		artist_label:               artist_label,
		controls_box:               controls_box,
		artist_box:                 artist_box,
		playpause_button:           playpause_button,
		previous_button:            previous_button,
		next_button:                next_button,
		playback_header:            playback_header,
		play_image:                 play_image,
		pause_image:                pause_image,
		status_icon:                status_icon,
		album_box:                  album_box,
		album_label:                album_label,
		album_art_image:            album_art_image,
		registered_action_handlers: make(map[Action]*list.List),
		buttonKeyMap:               buttonKeyMap,
		resourceManager:            resources.NewResourceManager(),
	}
}

// Initiates the GUI
func (this *GUI) Run() {

	this.main_window.Connect("destroy", func() {
		this.fireAction(ACTION_QUIT)
	})

	this.status_icon.Connect("activate", func() {
		log.Println("Status Icon clicked!")
		this.main_window.SetVisible(!this.main_window.GetVisible())
	})

	this.playpause_button.Connect("clicked", func() {
		this.fireAction(ACTION_PLAYPAUSE)
	})

	this.previous_button.Connect("clicked", func() {
		this.fireAction(ACTION_PREVIOUS)
	})

	this.next_button.Connect("clicked", func() {
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
	log.Printf("Extracting pressed key")
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
	log.Printf("Updating current song: %v", current_song)
	_, err := glib.IdleAdd(func() {
		this.title_label.SetText(current_song.Title)
		this.artist_label.SetText(current_song.Artist)

		if current_song.Album != "" {
			this.album_label.SetText(current_song.Album)
		} else {
			this.album_box.Hide()
		}

		this.main_window.SetTitle(fmt.Sprintf("psmpc: %s - %s", current_song.Artist, current_song.Title))
		this.status_icon.SetTooltipText(fmt.Sprintf("psmpc: %s - %s", current_song.Artist, current_song.Title))

		album_art_fp, err :=
			this.resourceManager.GetResourceAsFilePath(&resources.Track{current_song}, resources.ALBUM_ART)
		if err != nil {
			log.Println("Failed to get album art for %s", current_song)
		} else {
			this.album_art_image.SetFromFile(album_art_fp)
		}
	})
	if err != nil {
		log.Fatal("Failed to do glib.IdleAdd()")
	}
}

// Updates the GUI with the current MPD status
func (this *GUI) UpdateCurrentStatus(current_status *mpdinfo.Status) {
	log.Printf("Updating current status: %v", current_status)
	_, err := glib.IdleAdd(func() {
		switch current_status.State {

		case mpdinfo.STATE_STOPPED:
			this.controls_box.Hide()
			this.artist_box.Hide()
			this.album_box.Hide()
			this.title_label.SetText("Stopped")
			this.main_window.SetTitle("psmpc")
			this.status_icon.SetTooltipText("psmpc")

		case mpdinfo.STATE_PLAYING:
			this.controls_box.Show()
			this.artist_box.Show()
			this.album_box.Show()
			this.playpause_button.SetImage(this.pause_image)

		case mpdinfo.STATE_PAUSED:
			this.controls_box.Show()
			this.artist_box.Show()
			this.album_box.Show()
			this.playpause_button.SetImage(this.play_image)
		}
	})

	if err != nil {
		log.Fatal("Failed to do glib.IdleAdd()")
	}
}

// Fires the action specified by the given Action, passing the given arguments to all the
// subscribed handlers
func (this *GUI) fireAction(action_type Action, args ...interface{}) {
	log.Printf("Firing action %v", action_type)

	handlers, any := this.registered_action_handlers[action_type]

	if any == false {
		// None are registered
		log.Println("No action handlers found")
		return
	}

	log.Println("Handlers found ", handlers)

	for e := handlers.Front(); e != nil; e = e.Next() {
		log.Println("Executing handler", e.Value)
		handler := e.Value.(ActionHandler)
		go handler(args)
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
