package main

import (
	"log"
	"testing"
)

func TestSimplePars(t *testing.T) {
	in := "<a href=\"https://fsfd.fsdf\">"
	from := "<a href=\""
	to := "\">"

	out := simpleParse([]byte(in), from, to)

	if out == nil || string(out) == "https://fsfd.fsdf" {
		t.FailNow()
	}

	log.Println(string(out))
}

