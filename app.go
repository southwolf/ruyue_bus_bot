package main

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type SeatsResponse struct {
	code    int    `json:"code"`
	success string `json:"success"`
	msg     string `json:"msg"`
	count   int    `json:"count"`
	data    string `json:"data"`
}

var msg = strings.Builder{}
var startTime time.Time
var checkDisabled = false

func main() {
	startTime = time.Now()

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	addr := ":" + port
	http.HandleFunc("/", dashboard)

	log.Println("Server running on " + addr)
	go func() { log.Fatal(http.ListenAndServe(addr, nil)) }()
	for {
		go checkSwitch()
		go checkTickets()
		go keepAwake()
		time.Sleep(1 * time.Minute)
	}

}

func dashboard(w http.ResponseWriter, r *http.Request) {
	upTime := time.Since(startTime)
	w.Write([]byte("Running " + upTime.String() + "\n" + msg.String()))
}

func keepAwake() {
	appURL := "https://ruyue-bot.herokuapp.com/"
	get(appURL)
}

func checkSwitch() {
	// get last message
	botURL := "https://api.telegram.org/bot455106310:AAFvX2OlolvzLG4alNEncFAqh3XpRsU_zjM/getUpdates?offset=-1"
	result, err := get(botURL)
	if err != nil {
		log.Println(err)
	}
	fmt.Println(result.String())
	prev := checkDisabled
	checkDisabled = result.Get("result.0.message.text").String() == "0"
	if checkDisabled {
		msg.WriteString("Checking disabled on: " + time.Now().Format(time.Stamp))
		log.Println(msg.String())
	} else {
		msg.WriteString("Checking enabled on: " + time.Now().Format(time.Stamp))
		log.Println(msg.String())
	}
	if checkDisabled != prev {
		notify(msg.String())
	}
	return
}

func checkTickets() {
	if checkDisabled {
		return
	} else {
		msg = strings.Builder{}

		log.Println("Start checking on: " + time.Now().Format(time.Stamp))

		routesURL := "http://www.gzruyue.org.cn:11909/api/Product/ProductGetListByStationName?snm=%E4%BA%9A%E8%BF%90%E5%9F%8E"

		result, err := get(routesURL)
		if err != nil {
			log.Println(err)
		}

		routeNumber := ""
		result.Get("data.items").ForEach(func(key, route gjson.Result) bool {
			if route.Get("Routenm").String() == "亚运城->宏发广场" {
				routeNumber = route.Get("prolist").String()
				return false
			}
			return true
		})

		log.Println("Route Number: " + routeNumber)
		seatsURL := "http://www.gzruyue.org.cn:11909/api/Product/ProductDayArrayList?pid=" + routeNumber

		result, err = get(seatsURL)
		if err != nil {
			log.Println(err)
		}

		msg.WriteString(result.Get("data.pct").String() + " days to go:\n")

		isAvailable := false
		result.Get("data.items").ForEach(func(key, day gjson.Result) bool {
			day.Get("clsinf").ForEach(func(key, line gjson.Result) bool {
				date := day.Get("date").String()
				time := line.Get("clstm").String()
				seats := line.Get("seats")

				if seats.Int() > 0 {
					msg.WriteString(date + " " + time + " " + seats.String() + "\n")
					isAvailable = true
				}
				return true
			})
			return true
		})
		fmt.Print(msg.String())

		if isAvailable {
			notify(msg.String())
		}
	}
}

func notify(msg string) {
	botURL := "https://api.telegram.org/bot455106310:AAFvX2OlolvzLG4alNEncFAqh3XpRsU_zjM/sendMessage"
	msgJson := []byte(`{"chat_id":"552224197", "text":"` + msg + `"}`)
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, _ := http.NewRequest("POST", botURL, bytes.NewBuffer(msgJson))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	defer resp.Body.Close()
}

func get(URL string) (result gjson.Result, err error) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	req, _ := http.NewRequest("GET", URL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 11_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E216")

	response, err := client.Do(req)
	if err != nil {
		return gjson.Result{}, err
	}

	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		err = fmt.Errorf("API ERROR! %d: %s", response.StatusCode, body)
		return gjson.Result{}, err
	}

	result = gjson.Parse(string(body))

	return result, nil
}
