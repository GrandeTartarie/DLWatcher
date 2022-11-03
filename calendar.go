package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func (c *Calendar) GetActiveEvents() []Event {
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

func GetCalendar(sessKey string, client *Client) (*Calendar, error) {
	t := time.Now()

	ge := []Service{
		{
			Index:      0,
			MethodName: "core_calendar_get_calendar_monthly_view",
			Args: Args{
				Year:              t.Year(),
				Month:             int(t.Month()),
				CourseID:          1,
				CategoryID:        0,
				IncludeNavigation: true,
				Mini:              true,
				Day:               t.Day(),
			},
		},
	}

	out, err := json.Marshal(ge)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf(
			"https://dl.nure.ua/lib/ajax/service.php?sesskey=%s&info=%s",
			sessKey,
			ge[0].MethodName,
		),
		bytes.NewReader(out),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("accept", "*/*")
	req.Header.Set("user-agent", UAG)
	req.Header.Set("content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//fmt.Println(string(body))

	var calendar []Calendar
	if err = json.Unmarshal(body, &calendar); err != nil {
		return nil, err
	}

	return &calendar[0], nil
}
