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
	"strconv"
	"strings"
	"time"
)

var (
	UAG         = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Ubuntu Chromium/37.0.2062.94 Chrome/37.0.2062.94 Safari/537.36"
	EventWorker = NewEventor()
	ProxyURL    *url.URL
)

func main() {
	err := godotenv.Load("./.env")
	if err != nil {
		log.Panicln(err)
	}

	ReCheckEveryInMinutes, err = strconv.Atoi(os.Getenv("RECHECK_IN_MINUTES"))
	if err != nil {
		log.Panicln(err)
	}
	//fmt.Println(os.Getenv("DL_LOGIN"))

	//ProxyURL, err = url.Parse("http://127.0.0.1:8888")
	//if err != nil {
	//	log.Panicln(err)
	//}

	ticker := time.NewTicker(
		time.Duration(ReCheckEveryInMinutes) *
			time.Minute)
	quit := make(chan struct{})

	work()
	for {
		select {
		case <-ticker.C:
			work()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func work() {
	fmt.Printf("%s:\n", time.Now().Format(time.Stamp))
	// reset prev timers
	EventWorker.KillAll()

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

	fmt.Println("Authentication...")
	sessKey, err := auth(client)
	if err != nil {
		log.Panicln(err)
	}

	fmt.Println("Getting calendar...")
	calendar, err := getCalendar(sessKey, client)
	if err != nil {
		log.Panicln(err)
	}
	//fmt.Printf("%+v\n", calendar)

	events := calendar.GetActiveEvents()
	latency := int64(time.Minute.Seconds())

	//fmt.Printf("%+v\n", events)
	if len(events) == 0 {
		fmt.Println("Nothing to visit((")
	}

	EventWorker = VisitEvents(events, latency, client)
}

func VisitEvents(events []Event, latency int64, client *http.Client) *Eventor {
	eventor := NewEventor()

	for _, e := range events {
		sleep := e.TimeStart + latency - time.Now().Unix()

		if sleep < 0 {
			fmt.Printf(
				"Trying to visit '%s'...\n",
				e.Course.FullName,
			)
		} else {
			fmt.Printf(
				"'%s' will be visited after %d minutes\n",
				e.Course.FullName,
				sleep/int64(time.Minute.Seconds()),
			)
		}

		eventor.Add(func() {
			err := visitEvent(e, client)
			if err != nil {
				log.Println(err.Error())
			}

			fmt.Printf(
				"'%s' has been visited now\n",
				e.Course.FullName,
			)
		}, time.Second*time.Duration(sleep))
	}

	return eventor
}

func visitEvent(event Event, client *http.Client) error {
	resp, err := client.Get(event.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	from := "https://dl.nure.ua/mod/attendance/attendance.php?"
	u := simpleParse(body, from, "\"")

	s := from + string(u)
	_, err = client.Get(s)
	if err != nil {
		return err
	}

	return nil
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
