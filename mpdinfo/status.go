package mpdinfo

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
