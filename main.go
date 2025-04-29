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
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
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

type train_service struct {
	std      string
	etd      string
	plat     string
	dest     string
	operator string
	toc      string
}

func train(crs string, nr int) []train_service {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	const base_url string = "https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepartureBoard/"

	crs = strings.ToUpper(strings.TrimSpace(crs))

	params, err := format_params([]string{"numRows"},
		[]string{fmt.Sprintf("%v", nr)})

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

func refershTimes(mylabel_addr **widget.Label) {
	const desired_len = 10
	updated_times := train("RDG", (desired_len + 1))
	var updated_str string = ""
	for idx, val := range updated_times {
		if idx >= desired_len {
			break
		}
		updated_str += fmt.Sprintf("plat %s %s %s to %s expected %s", val.plat, val.toc, val.std, val.dest, val.etd) + "\n"
	}
	mylabel_obj := *mylabel_addr
	mylabel_obj.SetText(updated_str)
}

func tidyUp() {
	fmt.Println("exiting app, thank you for using Quick Train Times")
}

func main() {
	defer tidyUp()

	myapp := app.New()
	mywin := myapp.NewWindow("Quick Train Times")

	mywin.SetContent(widget.NewLabel("Welcome to Quick Train Times"))

	tt_label := widget.NewLabel("train time goes here")
	mywin.SetContent(tt_label)

	refershTimes(&tt_label)

	go func() {
		for range time.Tick(time.Minute) {
			refershTimes(&tt_label)
		}
	}()

	mywin.Show()
	myapp.Run()

}
