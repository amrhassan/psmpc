package gui

/*
#cgo pkg-config: gdk-3.0
#include <gdk/gdk.h>
*/
import "C"

import (
	"container/list"
	"fmt"
	"github.com/amrhassan/psmpc/logging"
	"github.com/amrhassan/psmpc/mpdinfo"
	"github.com/amrhassan/psmpc/resources"
	"github.com/conformal/gotk3/gdk"
	"github.com/conformal/gotk3/glib"
	"github.com/conformal/gotk3/gtk"
	"os"
	"unsafe"
)

var logger = logging.New("gui")

/*
 * The paths where the static resources are looked up from. The paths are tried in the order
 * they are listed in, and the first one that exists is used.
 */
var static_resource_file_paths = []string{
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
	elapsedProgressBar         *gtk.ProgressBar
	resourceManager            *resources.ResourceManager
	currentSong                *mpdinfo.CurrentSong
}

func error_panic(message string, err error) {
	panic(message + ": " + err.Error())
}

func path_exists(path string) bool {
	_, path_error := os.Stat(path)
	return path_error == nil || os.IsExist(path_error)
}

// A global cahe of static resources. Used by get_static_resource_path()
var staticResources = make(map[string]string)

func get_static_resource_path(resourceName string) string {

	entry, exists := staticResources[resourceName]
	if exists {
		return entry
	}

	for _, path := range static_resource_file_paths {
		full_path := path + "/gui/" + resourceName
		if path_exists(full_path) {
			logger.Info("Using %s file from: %s", resourceName, full_path)
			staticResources[resourceName] = full_path
			return full_path
		}
	}

	logger.Fatal("Can't find %s file", resourceName)
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

	err = builder.AddFromFile(get_static_resource_path("ui.glade"))
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
	main_window.SetIconFromFile(get_static_resource_path("icon.png"))

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
	status_icon, _ := gtk.StatusIconNewFromFile(get_static_resource_path("icon.png"))
	album_box := getGtkObject("album_box").(*gtk.Box)
	album_label := getGtkObject("album_label").(*gtk.Label)
	album_art_image := getGtkObject("album_art").(*gtk.Image)
	elapsed_progress_bar := getGtkObject("elapsed_progressbar").(*gtk.ProgressBar)

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
		elapsedProgressBar:         elapsed_progress_bar,
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
		logger.Debug("Status Icon clicked!")
		this.main_window.Show()

		this.main_window.SetUrgencyHint(true)
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
		logger.Debug("Key press: %v", key)

		action, mapped := this.buttonKeyMap[key.value]

		if mapped {
			this.fireAction(action)
		} else if key.value == 65307 {
			this.main_window.Hide()
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

func (this *key) String() string {
	return fmt.Sprintf("key{%d, %s}", this.value, this.representation)
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

	if this.currentSong != nil && *current_song == *(this.currentSong) {
		return
	}

	logger.Info("Updating current song: %v", current_song)
	this.currentSong = current_song

	executeInGlibLoop(func() {
		this.title_label.SetText(current_song.Title)
		this.artist_label.SetText(current_song.Artist)

		if current_song.Album != "" {
			this.album_label.SetText(current_song.Album)
		} else {
			this.album_box.Hide()
		}

		this.main_window.SetTitle(fmt.Sprintf("psmpc: %s - %s", current_song.Artist, current_song.Title))
		this.status_icon.SetTooltipText(fmt.Sprintf("psmpc: %s - %s", current_song.Artist, current_song.Title))
		this.album_art_image.SetFromFile(get_static_resource_path("album.png"))
	})

	go func() {
		album_art_fp, err :=
			this.resourceManager.GetResourceAsFilePath(&resources.Track{current_song}, resources.ALBUM_ART)
		if err != nil {
			logger.Warn("Failed to get album art for %s", current_song)
		} else {
			executeInGlibLoop(func() {
				this.album_art_image.SetFromFile(album_art_fp)
			})
		}
	}()
}

// The only thread-safe way to execute GTK-manipulating code.
func executeInGlibLoop(code func()) {
	_, err := glib.IdleAdd(code)
	if err != nil {
		logger.Fatal("Failed to do glib.IdleAdd()")
	}
}

// Updates the GUI with the current MPD status
func (this *GUI) UpdateCurrentStatus(current_status *mpdinfo.Status) {
	logger.Info("Updating current status: %v", current_status)

	executeInGlibLoop(func() {

		switch current_status.State {

		case mpdinfo.STATE_STOPPED:
			this.controls_box.Hide()
			this.artist_box.Hide()
			this.album_box.Hide()
			this.title_label.SetText("Stopped")
			this.main_window.SetTitle("psmpc")
			this.status_icon.SetTooltipText("psmpc")
			this.album_art_image.SetFromFile(get_static_resource_path("album.png"))
			this.currentSong = nil

		case mpdinfo.STATE_PLAYING:
			this.controls_box.Show()
			this.artist_box.Show()
			this.album_box.Show()
			this.playpause_button.SetImage(this.pause_image)
			this.elapsedProgressBar.SetFraction(current_status.SongProgress)

		case mpdinfo.STATE_PAUSED:
			this.controls_box.Show()
			this.artist_box.Show()
			this.album_box.Show()
			this.playpause_button.SetImage(this.play_image)
			this.elapsedProgressBar.SetFraction(current_status.SongProgress)
		}
	})
}

// Fires the action specified by the given Action, passing the given arguments to all the
// subscribed handlers
func (this *GUI) fireAction(action_type Action, args ...interface{}) {
	logger.Debug("Firing action %v", action_type)

	handlers, any := this.registered_action_handlers[action_type]

	if any == false {
		// None are registered
		logger.Debug("No action handlers found")
		return
	}

	logger.Debug("Handlers found ", handlers)

	for e := handlers.Front(); e != nil; e = e.Next() {
		logger.Debug("Executing handler", e.Value)
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
