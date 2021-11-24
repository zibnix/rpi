package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"time"
)

type SensorData struct {
	Temperature float64 `json:"temperature"`
	Humidity    float64 `json:"humidity"`
	Timestamp   string  `json:"timestamp"`
	Valid       bool    `json:"valid"`
}

// call Loop on a goroutine, it will use time.Sleep
// Loop returns immediately if the sensors is already looping
// call Stop to stop looping, sensors can be started and stopped
// unless the provided context.Context is done
type Sensors interface {
	Loop()
	Data() SensorData
	Stop()
}

type sensors struct {
	ctx        context.Context
	scriptFile string
	interval   time.Duration
	tempPin    uint

	reading bool
	latest  SensorData
	lk      sync.RWMutex

	wg   sync.WaitGroup
	quit chan struct{}
}

// script file should be a path to an executable that takes the GPIO pin number as a command line argument
// the GPIO pin number for a dht11 temp/humidity sensor is also provided in the tempPin argument
func NewSensors(ctx context.Context, scriptFile string, interval time.Duration, tempPin uint) Sensors {
	return &sensors{
		ctx:        ctx,
		scriptFile: scriptFile,
		interval:   interval,
		tempPin:    tempPin,
	}
}

func (s *sensors) Data() SensorData {
	s.lk.RLock()
	defer s.lk.RUnlock()

	return s.latest
}

func (s *sensors) Loop() {
	s.lk.Lock()
	if s.reading {
		s.lk.Unlock()
		return
	}
	s.reading = true
	s.quit = make(chan struct{})
	s.lk.Unlock()

	s.wg.Add(1)
	defer s.wg.Done()

	s.setData()

	for {
		timer := time.NewTimer(s.interval)

		select {
		case <-timer.C:
			s.setData()
		case <-s.ctx.Done():
			s.finished()
			return
		case <-s.quit:
			s.finished()
			return
		}
	}
}

func (s *sensors) setData() {
	data, err := s.read()
	if err != nil {
		log.Printf("Error while reading from sensor, will try again after interval: %v\n", err)
		return
	}

	log.Printf("Successful sensors reading: %v\n", data)

	s.lk.Lock()
	s.latest = data
	s.lk.Unlock()
}

func (s *sensors) finished() {
	s.lk.Lock()
	s.reading = false
	s.lk.Unlock()
}

func (s *sensors) Stop() {
	s.lk.Lock()
	if !s.reading {
		s.lk.Unlock()
		return
	}
	close(s.quit)
	s.lk.Unlock()

	s.wg.Wait()
}

type SensorRead struct {
	SensorData
	Error     bool `json:"error"`
	ErrorCode int  `json:"error_code"`
}

func (s *sensors) read() (SensorData, error) {
	var read SensorRead
	read.Timestamp = time.Now().Format(time.RFC822Z)

	out, err := exec.Command(s.scriptFile, fmt.Sprintf("%d", s.tempPin)).Output()
	if err != nil {
		return read.SensorData, err
	}

	if err := json.Unmarshal(out, &read); err != nil {
		return read.SensorData, err
	}

	if read.Error {
		return read.SensorData, fmt.Errorf("Error during read: %d", read.ErrorCode)
	}

	read.Valid = true

	return read.SensorData, nil
}
