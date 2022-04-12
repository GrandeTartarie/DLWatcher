package main

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
