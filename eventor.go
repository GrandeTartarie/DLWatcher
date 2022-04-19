package main

import (
	"time"
)

type Eventor struct {
	killers []Killer
}

type Killer chan struct{}

func NewEventor() *Eventor {
	return &Eventor{
		killers: []Killer{},
	}
}

func (ek *Eventor) Add(f func(), dur time.Duration) {
	killer := make(chan struct{}, 1)
	ek.killers = append(ek.killers, killer)

	t := time.NewTimer(dur)

	go func() {
		select {
		case <-killer:
			return
		case <-t.C:
			f()
		}
	}()
}

func (ek *Eventor) KillAll() {
	for _, k := range ek.killers {
		k <- struct{}{}
	}
	ek.killers = []Killer{}
}
