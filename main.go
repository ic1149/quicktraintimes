package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepBoardWithDetails/RDG
// https://api1.raildata.org.uk/1010-live-arrival-board-arr/LDBWS/api/20220120/GetArrBoardWithDetails/RDG
// https://api1.raildata.org.uk/1010-service-details1_2/LDBWS/api/20220120/GetServiceDetails/{serviceid}

func main() {
	const base_url string = "https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepBoardWithDetails/"
	const crs string = "RDG"
	var url string = base_url + strings.ToUpper(crs)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatal(err)
	}

	const dep_api_key string = ""
	req.Header.Set("x-apikey", dep_api_key)
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", body)

}
