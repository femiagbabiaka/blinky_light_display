package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

type displayPage struct {
	RowCount int
	Colsizes []int
	Rows     [][]string
}

var displayPagesString string = ""
var displayPages []displayPage = make([]displayPage, 0)
var currentPage displayPage = *new(displayPage)

func main() {

	//Attempt to read configuration file
	data, err := ioutil.ReadFile("conf.txt")
	if err == nil {
		configure(string(data), false)
	}

	//start managing the current url
	go manageCurrentUrl()

	//configuration REST endpoint
	http.HandleFunc("/configure/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method == "POST" {
			log.Println("Recieved post request.")
			configure(req.FormValue("list"), true)
			w.Write([]byte("Success!  http://localhost:3030/config.html"))
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Quit being a nudge, Spew."))
		}
	})

	//currentUrl REST endpoint
	http.HandleFunc("/current/", serveCurrentUrl)

	//allUrls REST endpoint
	http.HandleFunc("/all/", serveAllUrls)

	//static file server
	clientfs := http.FileServer(http.Dir("client"))
	http.Handle("/", clientfs)

	//Listen for connections and serve
	log.Println("Listening...")
	err2 := http.ListenAndServe(":3030", nil)
	if err2 != nil {
		log.Println(err.Error())
	}
}

func configure(list string, fromWeb bool) {
	log.Println("configuring")
	if fromWeb {
		data := []byte(list)
		ioutil.WriteFile("conf.txt", data, 0644)
	}
	displayPagesString = list
	displayPages = parseResourceList(displayPagesString)
}

func parseResourceList(resourceList string) []displayPage {

	log.Println("Parsing URLs.")
	resourceList = strings.Replace(resourceList, "\r", "", -1)

	rawPages := strings.Split(resourceList, "\n=\n")

	pages := make([]displayPage, len(rawPages))

	for i := range rawPages {
		thisPage := new(displayPage)
		rawRows := strings.Split(rawPages[i], "\n")
		thisPage.RowCount = len(rawRows)
		thisPage.Colsizes = make([]int, len(rawRows))
		thisPage.Rows = make([][]string, len(rawRows))
		for j := range rawRows {
			rawCols := strings.Split(rawRows[j], " | ")
			thisPage.Colsizes[j] = len(rawCols)
			thisPage.Rows[j] = make([]string, len(rawCols))
			for k := range rawCols {
				thisPage.Rows[j][k] = rawCols[k]
			}
		}
		pages[i] = *thisPage
	}
	return pages
}

func manageCurrentUrl() {
	for {
		if len(displayPages) > 0 {
			thisPage := displayPages[time.Time.Minute(time.Now())%len(displayPages)]
			currentPage = thisPage
			log.Printf("Set current url to %s\n", thisPage)
			time.Sleep(1 * time.Minute)
		} else {
			time.Sleep(1 * time.Second)
		}
	}
}

func serveCurrentUrl(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	urlJson, _ := json.Marshal(currentPage)
	w.Write(urlJson)

}

func serveAllUrls(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte(displayPagesString))
}
