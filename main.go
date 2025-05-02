package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
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

type quick_time struct {
	Start string
	End   string
	Org   string
	Dest  string
	Days  []int
}

type config struct {
	Dep_key     string
	Quick_times []any
}

func load_config() (string, []quick_time) {
	b, err := os.ReadFile("config.json")
	if err != nil {
		fmt.Println(err)
	}
	var c config
	err = json.Unmarshal(b, &c)
	if err != nil {
		fmt.Println(err)
	}

	quick_times := make([]quick_time, 0, len(c.Quick_times))
	for _, val := range c.Quick_times {
		this_map := val.(map[string]any)
		var this_struct quick_time
		this_struct.Start = this_map["start"].(string)
		this_struct.End = this_map["end"].(string)
		this_struct.Org = this_map["org"].(string)
		this_struct.Dest = this_map["dest"].(string)
		this_struct.Days = this_map["days"].([]int)

		quick_times = append(quick_times, this_struct)
	}
	return c.Dep_key, quick_times
}

func train(crs string, nr int) {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %s", err)
	// }

	const base_url string = "https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepartureBoard/"

	//crs = strings.ToUpper(strings.TrimSpace(crs))

	params, err := format_params([]string{"numRows"},
		[]string{fmt.Sprintf("%v", nr)})

	if err != nil {
		log.Fatal(err)
	}

	// var dep_api_key = os.Getenv("dep_key")
	dep_api_key, qts := load_config()
	var today int = int(time.Now().Weekday())
	correct_time := make([]quick_time, 0)
	for _, qt := range qts {
		if slices.Contains(qt.Days, today) {
			start, err := time.Parse("00:00", qt.Start)
			if err != nil {
				fmt.Println(err)
			}
			end, err := time.Parse("00:00", qt.End)
			now := time.Now()
			if now.After(start) && now.Before(end) {
				correct_time = append(correct_time, qt)
			}
		}
	}

	var url string = base_url + crs + params
	res := make([][]train_service, 0, len(correct_time))
	for range len(correct_time) {
		res = append(res, request(url, dep_api_key))
	}

}

// func refershTimes(mylabel_addr **widget.Label) {
// 	fmt.Println("refreshing train times")
// 	const desired_len = 10
// 	updated_times := train("RDG", (desired_len + 1))
// 	var updated_str string = ""
// 	for idx, val := range updated_times {
// 		if idx >= desired_len {
// 			break
// 		}
// 		updated_str += fmt.Sprintf("plat %s %s %s to %s expected %s", val.plat, val.toc, val.std, val.dest, val.etd) + "\n"
// 	}
// 	mylabel_obj := *mylabel_addr
// 	fyne.Do(func() { mylabel_obj.SetText(updated_str) })
// }

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
