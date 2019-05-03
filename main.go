package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

// For storing the project
var project = ""

// Set up redis connection pool
func init() {
	project = os.Getenv("crawl_project")
}

// scrapyDStatus is a type for handling scrapyd job statuses
type scrapyDStatus struct {
	Status  string `json:"status"`
	Pending []struct {
		ID     string `json:"id"`
		Spider string `json:"spider"`
	} `json:"pending"`
	Running []struct {
		ID        string `json:"id"`
		Spider    string `json:"spider"`
		StartTime string `json:"start_time"`
		Pid       int    `json:"pid"`
	} `json:"running"`
	Finished []struct {
		ID        string `json:"id"`
		Spider    string `json:"spider"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
	} `json:"finished"`
}

// logPanic is a wrapper for printing error mesages and crashing
func logPanic(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s \n", msg, err)
	}
}

// cancelAll cancels all the running spiders
func cancelAll() {
	// Init HTTP client
	client := &http.Client{}
	// Compose request
	req, err := http.NewRequest("GET",
		"http://localhost:6800/listjobs.json?project="+project, nil)
	logPanic(err, "Error constructing http request for scrapyd")
	// Make HTTP request
	res, err := client.Do(req)
	// If there's an error, stop here
	logPanic(err, "scrapyd error")
	// Parse response from scrapyd
	body, err := ioutil.ReadAll(res.Body)
	logPanic(err, "Error parsing response from scrapyd")
	// Init variable for storing the scrapy spider status
	var t scrapyDStatus
	// Store JSON result
	json.Unmarshal(body, &t)
	// For all running spiders
	for _, s := range t.Running {
		log.Println("CANCELLING " + s.ID)
		// Make new request to scrapyd
		req, err := http.NewRequest("POST",
			`http://localhost:6800/cancel.json?project=`+project+`&job=`+s.ID, nil)
		logPanic(err, "Error constructing http request for scrapyd")
		res, err = client.Do(req)
		logPanic(err, "Error from scrapyd response")
	}
}

func main() {
	flag.Parse()
	argAmnt := len(flag.Args())
	if argAmnt == 1 {
		switch flag.Args()[0] {
		case "cancel":
			cancelAll()
		default:
			break
		}
	} else {
		log.Print("Available operations:")
		log.Print("./spigo cancel")
	}
}
