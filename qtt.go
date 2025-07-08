package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"runtime"
	"slices"
	"sort"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// ----- global vars -----
var all_stations stations

var dayMapping = map[string]int{
	"Sun": 0,
	"Mon": 1,
	"Tue": 2,
	"Wed": 3,
	"Thu": 4,
	"Fri": 5,
	"Sat": 6,
}
var days = []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}

var qtt_cont_list []fyne.Container

var qts qtt

type station struct {
	Crs  string `json:"crs"`
	Name string `json:"Value"`
}

type stations struct {
	Version     string    `json:"version"`
	StationList []station `json:"StationList"`
}

type qtt struct {
	Quick_times []quick_time `json:"quick_times"`
	del_ids     []int
}

func (s *qtt) new_entry(new_qt quick_time) {
	existing_ids := []int{}
	for _, v := range s.Quick_times {
		existing_ids = append(existing_ids, v.Id)
	}
	// unique ids used to identify the correct qtt
	var new_id int

	for {
		new_id = rand.IntN(999) + 1 // 1-1000, arbitrary, doesn't matter
		if !slices.Contains(existing_ids, new_id) {
			break // unique id
		}
	}

	s.Quick_times = append(s.Quick_times, new_qt)
}

func (qts qtt) find_by_id(target_id int) (quick_time, error) {
	for _, val := range qts.Quick_times {
		if val.Id == target_id {
			return val, nil
		}
	} // linear search because unsorted

	return *new(quick_time), errors.New("not found")
}

func (qts qtt) check_exist(target_id int) bool {
	_, err := qts.find_by_id(target_id)
	if err != nil {
		return false // err, not found
	} else {
		return true // no err, found
	}
}

func (s *qtt) replace_by_id(target_id int, new_qt quick_time) {
	for i, v := range s.Quick_times {
		if v.Id == target_id {
			s.Quick_times[i] = new_qt
			break
		}
	}
}

func (s *qtt) del_by_id(target_id int) {
	new_slice := []quick_time{}
	for _, v := range s.Quick_times {
		if v.Id != target_id {
			new_slice = append(new_slice, v)
		}
	}
	s.Quick_times = new_slice
	s.del_ids = append(qts.del_ids, target_id)
}

func GetChosenDaysArray(selectedDays []string) []int {
	var chosenInts []int
	for _, dayStr := range selectedDays {
		if val, ok := dayMapping[dayStr]; ok {
			chosenInts = append(chosenInts, val)
		}
	}
	// Sort the array to ensure consistent order (e.g., [0, 1, 5] instead of [1, 5, 0])
	sort.Ints(chosenInts)
	return chosenInts
}

func time_validator(s string) error {
	if len(s) != 5 {
		return errors.New("incorrect length, should be 5 characters")
	} else if s[2] != ':' {
		return errors.New("not in correct time format")
	}
	// 15:39
	return nil
}

func crs_validator(s string) error {
	if len(s) != 3 {
		return errors.New("incorrect length, should be 3 letters")
	}

	for _, char := range s {
		if !unicode.IsLetter(char) || !unicode.IsUpper(char) {
			return errors.New("should be 3 uppercase English letters")
		}
	}

	_, err := crs_to_name(s)
	if err != nil {
		return fmt.Errorf("station not found in list, ver %v", all_stations.Version)
	} else {
		return nil // station found, no error
	}
}

func qtt_form(new bool, qt quick_time, mywin_addr *fyne.Window, rootURI fyne.URI) *fyne.Container {
	mywin_obj := *mywin_addr
	err := json.Unmarshal(resourceStationsJson.StaticContent, &all_stations)
	if err != nil {
		err_msg := dialog.NewError(err, mywin_obj)
		err_msg.Show()
	}

	entry_start := widget.NewEntry()
	entry_start.SetPlaceHolder("time in 24hr format e.g. 07:00")
	entry_start.Validator = time_validator

	entry_end := widget.NewEntry()
	entry_end.SetPlaceHolder("time in 24hr format e.g. 19:00")
	entry_end.Validator = time_validator

	entry_org := widget.NewEntry()
	entry_org.SetPlaceHolder("CRS code")
	entry_org.Validator = crs_validator

	entry_dest := widget.NewEntry()
	entry_dest.SetPlaceHolder("CRS code, or an asterisk (*) for any destination")
	entry_dest.Validator = func(s string) error {
		if s == "*" {
			return nil
		} else {
			err := crs_validator(s)
			return err
		}
	}

	checkDays := widget.NewCheckGroup(days, nil)

	if runtime.GOOS == "android" {
		checkDays.Horizontal = false // use vertical for mobile
	} else {
		checkDays.Horizontal = true
	}

	var id int
	if new {
		id = rand.IntN(999) + 1 // 1-999
		for {
			exist := qts.check_exist(id)
			if exist {
				id = rand.IntN(999) + 1 // 1-999
			} else {
				break
			}
		}
	} else {
		id = qt.Id
		entry_start.SetText(qt.Start)
		entry_end.SetText(qt.End)
		entry_org.SetText(qt.Org)
		entry_dest.SetText(qt.Dest)
		var selected_days []string
		for _, v := range qt.Days {
			selected_days = append(selected_days, days[v])
		}
		checkDays.SetSelected(selected_days)

	}

	form := &widget.Form{
		OnSubmit: func() {
			// save data
			var new_qt quick_time
			new_qt.Id = id
			new_qt.Start = entry_start.Text
			new_qt.End = entry_end.Text
			new_qt.Org = entry_org.Text
			new_qt.Dest = entry_dest.Text
			new_qt.Days = GetChosenDaysArray(checkDays.Selected)
			if qts.check_exist(id) {
				qts.replace_by_id(id, new_qt)
			} else {
				qts.new_entry(new_qt)
			}

			err := save_json(qts, "qtt.json", rootURI)
			if err != nil {
				err_msg := dialog.NewError(err, mywin_obj)
				err_msg.Show()
			} else {
				success_msg := dialog.NewInformation("Info", "entry saved successfully", mywin_obj)
				success_msg.Show()
			}

		},
		OnCancel: func() {
			// restore initial data
			cancel_msg := dialog.NewInformation("Info", "changes cancelled", mywin_obj)
			cancel_msg.Show()
			entry_start.SetText(qt.Start)
			entry_end.SetText(qt.End)
			entry_org.SetText(qt.Org)
			entry_dest.SetText(qt.Dest)
			var selected_days []string
			for _, v := range qt.Days {
				selected_days = append(selected_days, days[v])
			}
			checkDays.SetSelected(selected_days)
		},
		SubmitText: "Save",
		CancelText: "Cancel",
	}

	form.Append("Start time", entry_start)
	form.Append("End time", entry_end)
	form.Append("From station", entry_org)
	form.Append("To station", entry_dest)
	form.Append("Days", checkDays)

	del_button := widget.NewButtonWithIcon("", theme.DeleteIcon(), nil)

	del_confirm := dialog.NewConfirm("Deletion", "Are you sure you want to delete this entry?", func(b bool) {
		if b {
			qts.del_by_id(id)
			err := save_json(qts, "qtt.json", rootURI)
			if err != nil {
				err_msg := dialog.NewError(err, mywin_obj)
				err_msg.Show()
			} else {
				form.Hide()
				success_msg := dialog.NewInformation("Info", "entry deleted successfully", mywin_obj)
				success_msg.Show()
				del_button.Hide()
			}
		}
	}, mywin_obj)

	del_button.OnTapped = func() { del_confirm.Show() }

	form_border := container.NewBorder(nil, widget.NewLabel(" "), nil, del_button, form)

	return form_border
}

func qtt_init(mywin_addr *fyne.Window, rootURI fyne.URI) *container.Scroll {
	// GUI for qtt page
	mywin := *mywin_addr
	var err error
	_, qts, err = load_json("qtt.json", rootURI)
	if err != nil {
		dialog.NewError(err, mywin)
	}

	vb := container.NewVBox()

	for _, qt := range qts.Quick_times {
		qtt_cont_list = append(qtt_cont_list, *qtt_form(false, qt, mywin_addr, rootURI))
		vb.Add(&qtt_cont_list[len(qtt_cont_list)-1])
	}

	new_button := widget.NewButton("new entry", func() {
		qtt_cont_list = append(qtt_cont_list, *qtt_form(true, *new(quick_time), mywin_addr, rootURI))
		vb.Add(&qtt_cont_list[len(qtt_cont_list)-1])
	})

	main_border := container.NewBorder(nil, new_button, nil, nil, vb)
	return container.NewScroll(main_border)

}
