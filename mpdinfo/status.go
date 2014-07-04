package mpdinfo

import (
	"fmt"
)

type State string

const (
	STATE_PLAYING State = "play"
	STATE_STOPPED State = "stop"
	STATE_PAUSED  State = "pause"
)

type Status struct {
	Volume                          int
	Repeat, Random, Single, Consume bool
	State                           State
	SongProgress                    float64 // A fraction between 0.0 and 1.0
}

func (this *Status) String() string {
	return fmt.Sprintf(
		"Status{Volume: %d, Repeat: %v, Random: %v, Single: %v, Consume: %v, State: %v, SongProgress: %.2f}",
		this.Volume, this.Repeat, this.Random, this.Single, this.Consume, this.State, this.SongProgress,
	)
}
