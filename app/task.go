package app

import (
	"sync"

	"github.com/rs/zerolog/log"
)

type Task struct {
	wg sync.WaitGroup
}

func NewTask() *Task {
	return &Task{}
}

func (s *Task) Background(fn func()) {
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()

		defer func() {
			err := recover()
			if err != nil {
				log.Err(err.(error)).Stack().Send()
			}
		}()

		fn()
	}()
}

func (s *Task) Wait() {
	s.wg.Wait()
}
