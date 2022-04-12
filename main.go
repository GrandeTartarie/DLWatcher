package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	UAG = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/37.0.2062.94 Chrome/37.0.2062.94 Safari/537.36"
)

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Panicln(err)
	}
	//fmt.Println(os.Getenv("DL_LOGIN"))

	proxyUrl, err := url.Parse("http://127.0.0.1:8888")
	if err != nil {
		log.Panicln(err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Got error while creating cookie jar %s", err.Error())
	}

	client := &http.Client{
		Jar:       jar,
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyUrl)},
	}

	sessKey, err := auth(client)
	if err != nil {
		log.Panicln(err)
	}

	calendar, err := getCalendar(sessKey, client)
	if err != nil {
		log.Panicln(err)
	}
	fmt.Printf("%+v\n", calendar)
	dayN := calendar.Data.Date.YDay

	var events []Event
	for _, w := range calendar.Data.Weeks {
		for _, d := range w.Days {
			if d.YDay >= dayN {
				for _, e := range d.Events {
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

	wg := sync.WaitGroup{}
	wg.Add(len(events))
	latency := int64(time.Minute.Seconds())

	//fmt.Printf("%+v\n", events)
	if len(events) == 0 {
		fmt.Println("Nothing to visit((")
	}

	for _, event := range events {
		e := event
		go func() {
			defer wg.Done()
			sleep := e.TimeStart + latency - time.Now().Unix()

			fmt.Printf(
				"'%s' will be visited after %d minutes\n",
				e.Course.FullName,
				sleep/int64(time.Minute.Seconds()),
			)

			time.Sleep(time.Second * time.Duration(sleep))
			visitEvent(e, client)
			fmt.Printf(
				"'%s' has been visited now\n",
				e.Course.FullName,
			)
		}()
	}

	wg.Wait()

	fmt.Printf("Press Enter to exit...")
	fmt.Scanf("h")
}

func visitEvent(event Event, client *http.Client) {

}

func auth(client *http.Client) (string, error) {
	resp, err := client.Get("https://dl.nure.ua/login/index.php")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	loginToken := simpleParse(body, "name=\"logintoken\" value=\"", "\"")

	values := url.Values{
		"anchor":     {""},
		"logintoken": {string(loginToken)},
		"username":   {os.Getenv("DL_LOGIN")},
		"password":   {os.Getenv("DL_PASSWORD")},
	}

	req, err := http.NewRequest(
		"POST",
		"https://dl.nure.ua/login/index.php",
		strings.NewReader(values.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("accept", "*/*")
	req.Header.Set("user-agent", UAG)
	req.Header.Set("content-type", "application/x-www-form-urlencoded")

	resp1, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp1.Body.Close()

	body, err = ioutil.ReadAll(resp1.Body)
	if err != nil {
		return "", err
	}

	return string(simpleParse(body, "\"sesskey\":\"", "\"")), nil
}

func getCalendar(sessKey string, client *http.Client) (*Calendar, error) {
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
			"https://dl.nure.ua/lib/ajax/service.php?sesskey=%s&info=core_calendar_get_calendar_monthly_view",
			sessKey,
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

func simpleParse(in []byte, from string, to string) []byte {
	indexFrom := bytes.Index(in, []byte(from))
	shift := indexFrom + len(from)

	if indexFrom != -1 {
		indexTo := bytes.Index(in[shift:], []byte(to))

		if indexTo != -1 {
			return in[shift : shift+indexTo]
		}
	}

	return nil
}
