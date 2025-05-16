package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// this code file handles api

// https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepBoardWithDetails/RDG
// https://api1.raildata.org.uk/1010-live-arrival-board-arr/LDBWS/api/20220120/GetArrBoardWithDetails/RDG
// https://api1.raildata.org.uk/1010-service-details1_2/LDBWS/api/20220120/GetServiceDetails/{serviceid}

// ?sth=idk&thing=idk_either
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

func request(url, key string) []train_service {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("x-apikey", key) // put api key in header
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
	}

	body, err := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		if res.StatusCode > 299 {
			fmt.Printf("Response failed with status code: %d\n", res.StatusCode)
		} else {
			fmt.Printf("HTTP status code: %d", res.StatusCode)
		}
	}
	if err != nil {
		fmt.Println(err)
	}

	// fmt.Printf("%s", body)

	res_struct := return_data{}

	json.Unmarshal(fmt.Appendf(nil, "%s", body), &res_struct)

	services := make([]train_service, 0, len(res_struct.TrainServices))

	for _, val := range res_struct.TrainServices {
		thisService := val.(map[string]any)
		thisDest := thisService["destination"].([]any)
		thisDestInner := thisDest[0].(map[string]any)

		// fmt.Printf("%d: Platform %s for the %s (expected %s)\n", idx+1, thisService["platform"], thisService["std"], thisService["etd"])
		// fmt.Printf("%s service to %s\n\n", thisService["operator"], thisDestInner["locationName"])

		var new_service_struct train_service
		if thisService["platform"] != nil {
			new_service_struct.plat = thisService["platform"].(string)
		} else {
			new_service_struct.plat = "?"
		}
		// put data into defined structs
		new_service_struct.std = thisService["std"].(string)
		new_service_struct.etd = thisService["etd"].(string)
		new_service_struct.operator = thisService["operator"].(string)
		new_service_struct.dest = thisDestInner["crs"].(string)
		new_service_struct.toc = thisService["operatorCode"].(string)

		services = append(services, new_service_struct)
	}
	return services
}
