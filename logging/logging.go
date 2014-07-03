package logging

import (
	"fmt"
	"log"
	"os"
)

var maxNameLength = 0

// Returns the logger name padded to a unified width
func padName(name string) string {
	return fmt.Sprintf("%-"+fmt.Sprintf("%d", maxNameLength)+"s", name)
}

type Log struct {
	logger *log.Logger
	name   string
}

func New(name string) *Log {

	if len(name) > maxNameLength {
		maxNameLength = len(name)
	}

	return &Log{
		logger: log.New(os.Stderr, "", log.LstdFlags),
		name:   name,
	}
}

func (this *Log) Info(format string, params ...interface{}) {
	this.logger.Printf(padName(this.name)+" INFO    "+format, params...)
}

func (this *Log) Debug(format string, params ...interface{}) {
	this.logger.Printf(padName(this.name)+" DEBUG   "+format, params...)
}

func (this *Log) Warn(format string, params ...interface{}) {
	this.logger.Printf(padName(this.name)+" WARNING "+format, params...)
}

func (this *Log) Fatal(format string, params ...interface{}) {
	this.logger.Fatalf(padName(this.name)+" FATAL   "+format, params...)
}

func (this *Log) FatalError(err error) {
	this.Fatal(err.Error())
}
