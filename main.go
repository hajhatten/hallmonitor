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

	"github.com/gorilla/mux"
)

const (
	siteID      = "740049185"
	maxJourneys = 20
)

var (
	debug          = false
	apikeyStopSign = ""
)

func main() {

	flag.BoolVar(&debug, "d", false, "print debug messages")
	flag.Parse()

	apikeyStopSign = os.Getenv("RESROBOTAPIKEY")
	if apikeyStopSign == "" {
		log.Fatal("no apikey present, use env var RESROBOTAPIKEY to set it")
	}

	if debug {
		// Print result

		apiRespBody, err := readStopSignAPI()
		if err != nil {
			log.Fatal(err)
		}

		var slResponse slAPIResponse
		if err := json.Unmarshal(*apiRespBody, &slResponse); err != nil {
			log.Fatal(err)
		}

		result := sortAPIResponse(slResponse)

		fmt.Println("Bussar in mot stan: \n ")
		fmt.Println("Linje:\tDestination:\tAnkomst:\tOm:")
		printArrivals(result.ArrivalsInToCity)

		fmt.Println("\nBussar ut från stan: \n ")
		fmt.Println("Linje:\tDestination:\t\tAnkomst:\tOm:")
		printArrivals(result.ArrivalsOutOfCity)
	} else {
		router := mux.NewRouter()
		router.HandleFunc("/halltider", halltiderAPI)
		router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
		srv := &http.Server{
			Handler:      router,
			Addr:         ":3001",
			WriteTimeout: 15 * time.Second,
			ReadTimeout:  15 * time.Second,
		}
		log.Fatal(srv.ListenAndServe())
	}

}

func halltiderAPI(w http.ResponseWriter, r *http.Request) {

	apiRespBody, err := readStopSignAPI()
	if err != nil {
		log.Println(err)
		return
	}

	var slResponse slAPIResponse
	if err := json.Unmarshal(*apiRespBody, &slResponse); err != nil {
		log.Println(err)
		return
	}

	response, err := json.Marshal(sortAPIResponse(slResponse))
	if err != nil {
		log.Println(err)
		return
	}

	_, err = w.Write(response)
	if err != nil {
		log.Println(err)
	}
}

func printArrivals(arrivals []arrival) {
	for _, arrival := range arrivals {
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
		case "Alvik T-bana (Stockholm kn)":
			fmt.Printf("%s \tSpånga station \t\t%s \t%.0fm \n", arrival.TransportNumber, arrival.Time, math.Ceil(delta.Minutes()))
		case "Solna centrum T-bana":
			fmt.Printf("%s \tBlackebergs gård \t%s \t%.0fm \n", arrival.TransportNumber, arrival.Time, math.Ceil(delta.Minutes()))
		case "Tritonvägen (Sundbyberg kn)":
			fmt.Printf("%s \tBlackebergs gård \t%s \t%.0fm \n", arrival.TransportNumber, arrival.Time, math.Ceil(delta.Minutes()))
		}
	}
}

type slAPIResponse struct {
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

// Read resrobots hållplats-api
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

	if debug {
		if err := ioutil.WriteFile("result.json", body, 0777); err != nil {
			return nil, err
		}
	}

	return &body, nil
}

type apiResponse struct {
	ArrivalsInToCity  []arrival `json:"arrivalsInToCity"`
	ArrivalsOutOfCity []arrival `json:"arrivalsOutOfCity"`
}

// Sort response from resrobot api in two different directions based on origin
func sortAPIResponse(response slAPIResponse) (result apiResponse) {

	for _, arrival := range response.Arrivals {
		if debug {
			log.Println(arrival.Product, arrival.Origin)
		}
		switch arrival.Origin {
		case "Spånga station (Stockholm kn)":
			result.ArrivalsInToCity = append(result.ArrivalsInToCity, arrival)
		case "Blackebergs gård (Stockholm kn)":
			result.ArrivalsInToCity = append(result.ArrivalsInToCity, arrival)
		case "Alvik T-bana (Stockholm kn)":
			result.ArrivalsOutOfCity = append(result.ArrivalsOutOfCity, arrival)
		case "Solna centrum T-bana":
			result.ArrivalsOutOfCity = append(result.ArrivalsOutOfCity, arrival)
		case "Tritonvägen (Sundbyberg kn)":
			result.ArrivalsOutOfCity = append(result.ArrivalsOutOfCity, arrival)
		}
	}

	return
}
