package main

import "time"

type Service struct {
	Index      int    `json:"index"`
	MethodName string `json:"methodname"`
	Args       Args   `json:"args"`
}

type Args struct {
	Year              int  `json:"year"`
	Month             int  `json:"month"`
	CourseID          int  `json:"courseid"`
	CategoryID        int  `json:"categoryid"`
	IncludeNavigation bool `json:"includenavigation"`
	Mini              bool `json:"mini"`
	Day               int  `json:"day"`
}

type TimersType []*time.Timer

func (tt TimersType) Close() {
	for _, t := range tt {
		if !t.Stop() {
			<-t.C
		}
	}
}
