package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

// https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepBoardWithDetails/RDG
// https://api1.raildata.org.uk/1010-live-arrival-board-arr/LDBWS/api/20220120/GetArrBoardWithDetails/RDG
// https://api1.raildata.org.uk/1010-service-details1_2/LDBWS/api/20220120/GetServiceDetails/{serviceid}

func format_params(param_list []string, val_list []string) (string, error) {
	if len(param_list) != len(val_list) {
		return "", errors.New("not same number of parameter names and values")
	} else if len(param_list) < 1 { // no param or val
		return "", nil // empty str, but not error
	} else {
		var param_string string = "?" + param_list[0] + "=" + val_list[0]
		for idx, val := range param_list[1:] {
			param_string += "&" + val + "=" + val_list[idx]
		}
		return param_string, nil
	}

}

type return_data struct {
	TrainServices []any `json:"trainServices"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	const base_url string = "https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepartureBoard/"
	var crs string
	fmt.Print("CRS code: ")
	fmt.Scanln(&crs)
	crs = strings.ToUpper(strings.TrimSpace(crs))

	params, err := format_params([]string{"numRows"},
		[]string{"11"})

	if err != nil {
		log.Fatal(err)
	}

	var url string = base_url + crs + params

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		log.Fatal(err)
	}

	var dep_api_key string = os.Getenv("dep_key")
	req.Header.Set("x-apikey", dep_api_key)
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	body, err := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		if res.StatusCode > 299 {
			log.Fatalf("Response failed with status code: %d", res.StatusCode)
		} else {
			fmt.Printf("HTTP status code: %d", res.StatusCode)
		}
	}
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%s", body)

	res_struct := return_data{}

	json.Unmarshal(fmt.Appendf(nil, "%s", body), &res_struct)

	for idx, val := range res_struct.TrainServices {
		thisService := val.(map[string]any)
		thisDest := thisService["destination"].([]any)
		thisDestInner := thisDest[0].(map[string]any)

		fmt.Printf("%d: Platform %s for the %s (expected %s)\n", idx+1, thisService["platform"], thisService["std"], thisService["etd"])
		fmt.Printf("%s service to %s\n\n", thisService["operator"], thisDestInner["locationName"])
	}

}
