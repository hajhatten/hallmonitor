package main

// Run with command: RESROBOTAPIKEY=<APIKEY> while true; do clear; go run main.go; sleep 300; done

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"time"
)

const (
	siteID      = "740049185"
	maxJourneys = 20
)

var (
	debug          = false
	apikeyStopSign = ""
)

type apiresponse struct {
	Arrivals []arrival `json:"Arrival"`
}

type arrival struct {
	Product           product `json:"Product"`
	Stops             stops   `json:"Stops"`
	Name              string  `json:"name"`
	Stop              string  `json:"stop"`
	StopID            string  `json:"stopid"`
	StopExtID         string  `json:"stopExtId"`
	Time              string  `json:"time"`
	Date              string  `json:"date"`
	Origin            string  `json:"origin"`
	TransportNumber   string  `json:"transportNumber"`
	TransportCategory string  `json:"transportCategory"`
}

type product struct {
	Name         string `json:"name"`
	Num          string `json:"num"`
	CatCode      string `json:"catCode"`
	CatOutS      string `json:"catOutS"`
	CatOutL      string `json:"catOutL"`
	OperatorCode string `json:"operatorCode"`
	Operator     string `json:"operator"`
	OperatorURL  string `json:"operatorUrl"`
}

type stops struct {
	Stop []stop `json:"Stop"`
}

type stop struct {
	Name     string `json:"name"`
	ID       string `json:"id"`
	ExtID    string `json:"extId"`
	RouteIDx int64  `json:"routeIdx"`
	Lon      int64  `json:"lon"`
	Lat      int64  `json:"lat"`
	DepTime  string `json:"depTime"`
	DepDate  string `json:"depDate"`
}

func main() {

	flag.BoolVar(&debug, "d", false, "print debug messages")
	flag.Parse()

	apikeyStopSign = os.Getenv("RESROBOTAPIKEY")
	if apikeyStopSign == "" {
		log.Fatal("no apikey present, use env var RESROBOTAPIKEY to set it")
	}

	apiRespBody, err := readStopSignAPI()
	if err != nil {
		log.Fatal(err)
	}

	var response apiresponse
	if err := json.Unmarshal(*apiRespBody, &response); err != nil {
		log.Fatal(err)
	}

	// Sort response in two different directions

	var resultInToCity []arrival
	var resultOutOfCity []arrival

	for _, arrival := range response.Arrivals {
		switch arrival.Origin {
		case "Spånga station (Stockholm kn)":
			resultInToCity = append(resultInToCity, arrival)
		case "Blackebergs gård (Stockholm kn)":
			resultInToCity = append(resultInToCity, arrival)
		case "Alvik T-bana (Stockholm kn)":
			resultOutOfCity = append(resultOutOfCity, arrival)
		case "Solna centrum T-bana":
			resultOutOfCity = append(resultOutOfCity, arrival)
		}
	}

	// Print result

	fmt.Println("Bussar in mot stan: \n ")
	fmt.Println("Linje:\tDestination:\tAnkomst:\tOm:")

	for _, arrival := range resultInToCity {
		deptime, err := time.Parse("2006-01-02T15:04:05", fmt.Sprintf("%sT%s", arrival.Date, arrival.Time))
		if err != nil {
			log.Fatal(err)
		}
		delta := deptime.Sub(time.Now().Add(time.Hour * 2)).Truncate(time.Minute)

		switch arrival.Origin {
		case "Spånga station (Stockholm kn)":
			fmt.Printf("%s \tAlvik \t\t%s\t%.0fm \n", arrival.TransportNumber, arrival.Time, math.Ceil(delta.Minutes()))
		case "Blackebergs gård (Stockholm kn)":

			fmt.Printf("%s \tSolna centrum \t%s \t%.0fm \n", arrival.TransportNumber, arrival.Time, math.Ceil(delta.Minutes()))
		}
	}

	fmt.Println("\n \nBussar ut från stan: \n ")
	fmt.Println("Linje:\tDestination:\t\tAnkomst:\tOm:")
	for _, arrival := range resultOutOfCity {
		deptime, err := time.Parse("2006-01-02T15:04:05", fmt.Sprintf("%sT%s", arrival.Date, arrival.Time))
		if err != nil {
			log.Fatal(err)
		}
		delta := deptime.Sub(time.Now().Add(time.Hour * 2)).Truncate(time.Minute)

		switch arrival.Origin {
		case "Alvik T-bana (Stockholm kn)":
			fmt.Printf("%s \tSpånga station \t\t%s \t%.0fm \n", arrival.TransportNumber, arrival.Time, math.Ceil(delta.Minutes()))
		case "Solna centrum T-bana":
			fmt.Printf("%s \tBlackebergs gård \t%s \t%.0fm \n", arrival.TransportNumber, arrival.Time, math.Ceil(delta.Minutes()))
		}
	}
}

func readStopSignAPI() (*[]byte, error) {

	client := http.Client{
		Timeout: time.Second * 20,
	}

	url := fmt.Sprintf("https://api.resrobot.se/v2/arrivalBoard.json?key=%s&id=%s&maxJourneys=%d", apikeyStopSign, siteID, maxJourneys)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept-Encoding", "gzip, deflate")

	if debug {
		log.Println("==> Calling: ", url)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &body, nil
}
