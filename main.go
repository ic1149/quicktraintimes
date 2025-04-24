package main

import (
	"encoding/json"
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

func format_params(param_list []string, val_list []string) string {
	if len(param_list) != len(val_list) {
		log.Fatal("Invalid parameters: Not same number of parameter names and values")
		return ""
	} else {
		var param_string string = "?" + param_list[0] + "=" + val_list[0]
		for idx, val := range param_list[1:] {
			param_string += "&" + val + "=" + val_list[idx]
		}
		return param_string
	}

}

type return_data struct {
	TrainServices []interface{} `json:"trainServices"`
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	const base_url string = "https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepartureBoard/"
	const crs string = "RDG"
	var params = format_params([]string{"numRows"},
		[]string{"2"})

	var url string = base_url + strings.ToUpper(crs) + params

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

	if res.StatusCode > 299 {
		log.Fatalf("Response failed with status code: %d and\nbody: %s\n", res.StatusCode, body)
	}
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf("%s", body)

	res_struct := return_data{}

	json.Unmarshal([]byte(fmt.Sprintf("%s", body)), &res_struct)

	firstService := res_struct.TrainServices[0].(map[string]interface{})

	dest := firstService["destination"].([]interface{})

	dest_inner := dest[0].(map[string]interface{})

	fmt.Println(dest_inner["crs"])

}
