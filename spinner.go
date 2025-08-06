package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Spinner struct {
	frames []string
	pos    int
	stop   chan bool
	done   chan bool
}

func NewSpinner() *Spinner {
	return &Spinner{
		frames: []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		pos:    0,
		stop:   make(chan bool),
		done:   make(chan bool),
	}
}

func (s *Spinner) Start(message string) {
	// Handle interrupt signals to clean up spinner
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-s.stop:
				s.done <- true
				return
			case <-sigChan:
				s.Stop()
				os.Exit(1)
			default:
				fmt.Printf("\r%s %s", s.frames[s.pos], message)
				s.pos = (s.pos + 1) % len(s.frames)
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.stop <- true
	<-s.done
	fmt.Print("\r\033[K") // Clear the line
}

func (s *Spinner) UpdateMessage(message string) {
	fmt.Printf("\r%s %s", s.frames[s.pos], message)
}

// WithSpinner executes a function while showing a loading spinner
func WithSpinner(message string, fn func() error) error {
	spinner := NewSpinner()
	spinner.Start(message)
	
	// Execute the function in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- fn()
	}()
	
	// Wait for the function to complete
	err := <-errChan
	spinner.Stop()
	return err
} 