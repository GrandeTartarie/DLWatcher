package main

import (
	"fmt"
	"time"
)

type Calendar struct {
	Data Data `json:"data"`
}

type Data struct {
	Date  Date   `json:"date"`
	Weeks []Week `json:"weeks"`
}

type Date struct {
	YDay int `json:"yday"`
}

type Week struct {
	Days []Day `json:"days"`
}

type Day struct {
	Events    []Event `json:"events"`
	YDay      int     `json:"yday"`
	HasEvents bool    `json:"hasevents"`
}

type Event struct {
	IsActionEvent bool   `json:"isactionevent"`
	URL           string `json:"url"`
	TimeStart     int64  `json:"timestart"`
	TimeDuration  int64  `json:"timeduration"`
	Course        Course `json:"course"`
}

type Course struct {
	FullName string `json:"fullname"`
}

func (c Calendar) GetActiveEvents() []Event {
	var events []Event

	for _, w := range c.Data.Weeks {
		for _, d := range w.Days {
			if !d.HasEvents {
				continue
			}

			if d.YDay >= c.Data.Date.YDay {
				for _, e := range d.Events {
					if !e.IsActionEvent || e.TimeDuration == 0 {
						continue
					}

					if time.Now().Unix() >
						e.TimeStart+e.TimeDuration {
						fmt.Printf(
							"You've skipped event: '%s' at %s\n",
							e.Course.FullName,
							time.Unix(e.TimeStart, 0).
								Format(time.UnixDate),
						)
						continue
					}
					events = append(events, e)
				}
			}
		}
	}

	return events
}
