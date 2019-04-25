package main

import "time"

type EventResponse struct {
	Name        string
	Description string
	Comics      []*Comic
}

type Comic struct {
	Title string
	URL   string
	Image string
	Date  time.Time
}

type Event struct {
	Title string
	ID    int
}
