package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// this code file handles api

type wtapi struct {
	Offset string `json:"dst_offset"`
}

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
			param_string += "&" + val + "=" + val_list[idx+1]
		}
		return param_string, nil
	}

}

func request(url, key string) ([]train_service, error) {
	if key == "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx" {
		return nil, nil // default key, don't even bother sending request
	}
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, err
	}

	req.Header.Set("x-apikey", key) // put api key in header
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		if res.StatusCode > 299 {
			return nil, fmt.Errorf("HTTP Error Code %d", res.StatusCode)
		} else {
			return nil, fmt.Errorf("HTTP status code %d", res.StatusCode)
		}
	}
	if err != nil {
		return nil, err
	}

	res_struct := return_data{}

	json.Unmarshal(fmt.Appendf(nil, "%s", body), &res_struct)
	services := make([]train_service, 0, len(res_struct.TrainServices))

	for _, val := range res_struct.TrainServices {
		thisService := val.(map[string]any)
		thisDest := thisService["destination"].([]any)
		thisDestInner := thisDest[0].(map[string]any)

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
	return services, nil
}

func getLastSundayOfMonth(year int, month time.Month) time.Time {
	// Get the first day of the *next* month.
	// For example, if month is March, this gets April 1st.
	firstDayOfNextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.UTC)

	// Go back one day to get the last day of the *current* month.
	// For example, if firstDayOfNextMonth is April 1st, this gets March 31st.
	lastDayOfCurrentMonth := firstDayOfNextMonth.AddDate(0, 0, -1)

	// Iterate backwards from the last day of the month until it's a Sunday.
	currentDay := lastDayOfCurrentMonth
	for currentDay.Weekday() != time.Sunday {
		currentDay = currentDay.AddDate(0, 0, -1)
	}
	return currentDay
}

// IsUKUsingSummerTime determines if the UK is currently observing Summer Time (DST).
// DST in the UK is between 1:00 AM (GMT) on the last Sunday in March
// and 1:00 AM (GMT) on the last Sunday in October.
// The function uses the current UTC time for its check.
func IsUKUsingSummerTime() bool {
	// Get the current time in UTC (GMT) as the DST rules are specified in GMT.
	now := time.Now().UTC()

	year := now.Year()

	// Calculate the start of DST for the current year.
	// It's the last Sunday in March at 1:00 AM GMT.
	lastSundayMarch := getLastSundayOfMonth(year, time.March)
	dstStart := time.Date(year, lastSundayMarch.Month(), lastSundayMarch.Day(), 1, 0, 0, 0, time.UTC)

	// Calculate the end of DST for the current year.
	// It's the last Sunday in October at 1:00 AM GMT.
	lastSundayOctober := getLastSundayOfMonth(year, time.October)
	dstEnd := time.Date(year, lastSundayOctober.Month(), lastSundayOctober.Day(), 1, 0, 0, 0, time.UTC)

	// Check if the current time falls within the DST period.
	// The period is inclusive at the start and exclusive at the end: [dstStart, dstEnd)
	// This means 'now' must be greater than or equal to dstStart, AND
	// 'now' must be strictly less than dstEnd.
	return !now.Before(dstStart) && now.Before(dstEnd)
}
