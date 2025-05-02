package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func request(url, key string) []train_service {
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("x-apikey", key)
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
		new_service_struct.std = thisService["std"].(string)
		new_service_struct.etd = thisService["etd"].(string)
		new_service_struct.operator = thisService["operator"].(string)
		new_service_struct.dest = thisDestInner["locationName"].(string)
		new_service_struct.toc = thisService["operatorCode"].(string)

		services = append(services, new_service_struct)
	}
	return services
}
