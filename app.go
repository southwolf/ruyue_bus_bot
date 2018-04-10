package main

import (
	"bytes"
	"fmt"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"os"
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

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}
	addr := ":" + port
	http.HandleFunc("/", dashboard)
	log.Println("Server running on " + addr)
	log.Fatal(http.ListenAndServe(addr, nil))
	startTime = time.Now()
	for {
		time.Sleep(1 * time.Minute)
		go check()
	}


}

func dashboard(w http.ResponseWriter, r *http.Request) {
	upTime := time.Since(startTime)
	w.Write([]byte("Running " + upTime.String() + "\n" + msg.String()))
}

func check() {
	msg = strings.Builder{}

	log.Println("Start checking on: " + time.Now().Format(time.Stamp))
	client := &http.Client{
		Timeout: time.Second * 5,
	}

	seatsURL := "http://www.gzruyue.org.cn:11909/api/Product/ProductDayArrayList?pid=4854418523974249131"

	req, _ := http.NewRequest("GET", seatsURL, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (iPhone; CPU iPhone OS 11_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E216")

	response, err := client.Do(req)
	if err != nil {
		log.Print(err)
		return
	}

	body, _ := ioutil.ReadAll(response.Body)
	if response.StatusCode != 200 {
		log.Printf("%d: %s", response.StatusCode, body)
	}

	result := gjson.Parse(string(body))
	msg.WriteString(result.Get("data").Get("pct").String() + " days to go:\n")

	isAvailable := false
	result.Get("data").Get("items").ForEach(func(key, day gjson.Result) bool {
		day.Get("clsinf").ForEach(func(key, line gjson.Result) bool {
			date := day.Get("date").String()
			time := line.Get("clstm").String()
			seats := line.Get("seats")

			msg.WriteString(date + " " + time + " " + seats.String() + "\n")
			if seats.Int() > 0 {
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
