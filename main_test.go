package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"strconv"
	"testing"
)

func TestSimplePars(t *testing.T) {
	in := "<a href=\"https://example.url\">"
	from := "<a href=\""
	to := "\">"

	out := simpleParse([]byte(in), from, to)

	if out == nil || string(out) == "https://example.url" {
		t.FailNow()
	}

	log.Println(string(out))
}

func TestCalendar_GetActiveEvents(t *testing.T) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Panicln(err)
	}

	ReCheckEveryInMinutes, err = strconv.Atoi(os.Getenv("RECHECK_IN_MINUTES"))
	if err != nil {
		log.Panicln(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}

	tr := &http.Transport{}
	if ProxyURL != nil {
		tr.Proxy = http.ProxyURL(ProxyURL)
	}

	client := &http.Client{
		Jar:       jar,
		Transport: tr,
	}

	sessKey, err := auth(client)
	if err != nil {
		log.Panicln(err)
	}

	calendar, err := GetCalendar(sessKey, client)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Printf("%+v\n", calendar)
}
