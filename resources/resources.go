/*
Management and handling for music resources.

A resource is an artifact that is fetchable from a remote service provider of that
type of resource for a specified music Track.
*/
package resources

import (
	"github.com/amrhassan/psmpc/mpdinfo"
	"io"
)

type Track struct {
	*mpdinfo.CurrentSong
}

type ResourceType int

const (
	ALBUM_ART ResourceType = iota
	LYRICS    ResourceType = iota
)

// A provider for a resource type
type ResourceProvider interface {

	// This should return the type which this provider provides
	GetType() ResourceProvider

	// This should return the binary version of the resource
	GetResource(track *Track) (io.ReadCloser, error)
}
