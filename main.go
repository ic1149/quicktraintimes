package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"time"
	"unicode"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

//go:generate go run fyne.io/tools/cmd/fyne@latest bundle -o bundled.go FyneApp.toml
//go:generate go run fyne.io/tools/cmd/fyne@latest bundle -o bundled.go -append stations.json
//go:generate go run fyne.io/tools/cmd/fyne@latest bundle -o bundled.go -append toc.json

// longest station CRS ['BHA', 'CGT', 'EDA', 'LCB', 'STI', 'SWB', 'SWC', 'WER', 'WHH', 'XWX', 'ZHS']
// longest station names
// [
// 'Bournemouth Hurn Airport (Bus)',
// 'Catterick Garrison Tesco (Bus)',
// 'Edinburgh Airport (Bus or Tram',
// 'Lochboisdale, South Uist (Bus)',
// 'Stadium of Light (T & W Metro)',
// 'Swaffham Tourist Inf Ctr (Bus)',
// 'Swinderby A46 Roundabout (Bus)',
// 'Wedgwood Old Road Bridge (Bus)',
// 'Whitehill Prince of Wales (Bus',
// 'Wallingford Market Place (Bus)',
// 'Kensington High St Underground'
// ]
// all with 30 characters, seems like it cant store more and truncates
// Edinburgh... and Whitehill... has missing closing brackets

// catch the api
type return_data struct {
	TrainServices []any `json:"trainServices"`
}

// needed data for each train service
type train_service struct {
	std      string
	etd      string
	plat     string
	dest     string
	operator string
	toc      string
}

// catch quick time json settings
type quick_time struct {
	Id    int    `json:"id"`
	Start string `json:"start"`
	End   string `json:"end"`
	Org   string `json:"org"`
	Dest  string `json:"dest"`
	Days  []int  `json:"days"`
}

type metadata struct {
	Details met_details
}

type met_details struct {
	Version string
}

type TableConfig struct {
	Data               [][]string // The actual data for the table cells
	ColHeaderTexts     []string   // Texts for the column headers
	RowHeaderTexts     []string   // Texts for the row headers
	CellTemplateText   string     // Placeholder text for data cell templates
	HeaderTemplateText string     // Placeholder text for header cell templates
	CornerHeaderText   string     // Text for the top-left corner header cell
}

type tocs struct {
	Version string `json:"version"`
	TOCList []toc  `json:"TOCList"`
}

type toc struct {
	Toc  string `json:"toc"`
	Name string `json:"Value"`
}

func NewTableConfig(data [][]string, colHeaders []string, rowHeaders []string) *TableConfig {
	return &TableConfig{
		Data:               data,
		ColHeaderTexts:     colHeaders,
		RowHeaderTexts:     rowHeaders,
		CellTemplateText:   "Cell Data",   // Default placeholder for data cells
		HeaderTemplateText: "Header Info", // Default placeholder for header cells
		CornerHeaderText:   "",            // Default for corner (often empty or "No.")
	}
}

func (tc *TableConfig) BuildTable(mywin_addr *fyne.Window) *widget.Table {
	// Basic validation for data
	if tc.Data == nil {
		// log.Println("Warning: TableConfig.Data is nil. Creating an empty table.")
		tc.Data = [][]string{} // Ensure Data is not nil to prevent panics later
	}

	// DataFunc: Returns the number of rows and columns in the table.
	// This is crucial for the table to know how many rows to expect.
	dataFunc := func() (int, int) {
		if len(tc.Data) == 0 {
			return 0, 0 // No data, so 0 rows and 0 columns.
		}
		// The number of rows is the length of the outer slice.
		// The number of columns is assumed from the length of the first row.
		// Ensure your data is consistent (all rows have the same number of columns).
		return len(tc.Data), len(tc.Data[0])
	}

	// CreateCellFunc: Called once to create a template fyne.CanvasObject for data cells.
	createCellFunc := func() fyne.CanvasObject {
		return widget.NewLabel(tc.CellTemplateText)
	}

	// UpdateCellFunc: Called to update the content of a data cell.
	updateCellFunc := func(id widget.TableCellID, cell fyne.CanvasObject) {
		label := cell.(*widget.Label)
		// Protect against out-of-bounds access to tc.Data
		// This ensures that if data is missing for a cell, it defaults to empty.
		if id.Row >= 0 && id.Row < len(tc.Data) &&
			id.Col >= 0 && id.Col < len(tc.Data[id.Row]) {
			label.SetText(tc.Data[id.Row][id.Col])
		} else {
			label.SetText("") // Default to empty if data is out of bounds
		}
	}

	// Create the table using widget.NewTable.
	table := widget.NewTable(dataFunc, createCellFunc, updateCellFunc)

	// Manually enable header visibility.
	// widget.NewTableWithHeaders would do this, along with setting sticky headers.
	table.ShowHeaderRow = true
	table.ShowHeaderColumn = true

	// Customize the header creation and update logic.
	// CreateHeader is called to create a template for header cells.
	table.CreateHeader = func() fyne.CanvasObject {
		return widget.NewLabel(tc.HeaderTemplateText)
	}

	// UpdateHeader is called to set the content of each specific header cell.
	table.UpdateHeader = func(id widget.TableCellID, template fyne.CanvasObject) {
		label := template.(*widget.Label)

		if id.Row == -1 && id.Col == -1 { // Corner cell
			label.SetText(tc.CornerHeaderText)
		} else if id.Row == -1 { // Column headers
			if id.Col >= 0 && id.Col < len(tc.ColHeaderTexts) {
				label.SetText(tc.ColHeaderTexts[id.Col])
			} else {
				label.SetText("") // Fallback if no header text defined
			}
		} else if id.Col == -1 { // Row headers
			if id.Row >= 0 && id.Row < len(tc.RowHeaderTexts) {
				label.SetText(tc.RowHeaderTexts[id.Row])
			} else {
				label.SetText("") // Fallback if no header text defined
			}
		}
	}
	table.OnSelected = func(pos widget.TableCellID) {
		cell_data := tc.Data[pos.Row][pos.Col]
		err := crs_validator(cell_data)
		if err == nil {
			// is valid crs cell
			mywin_obj := *mywin_addr
			find_name, err := crs_to_name(cell_data)
			if err != nil {
				dialog.ShowError(err, mywin_obj)
				return
			}
			dialog.ShowInformation("Station Name", find_name, mywin_obj)
			return
		} else if len(cell_data) == 2 {
			is_upper := true
			for _, char := range cell_data {
				if !unicode.IsLetter(char) || !unicode.IsUpper(char) {
					is_upper = false
					break
				}
			}
			if is_upper {
				// is valid TOC code
				mywin_obj := *mywin_addr
				low := 0
				high := len(toc_names.TOCList)
				for high > low {
					middle := (low + high) / 2 // floor
					if toc_names.TOCList[middle].Toc == cell_data {
						dialog.ShowInformation("TOC Name", toc_names.TOCList[middle].Name, mywin_obj)
						return
					} else if cell_data > toc_names.TOCList[middle].Toc {
						low = middle + 1 // select right
					} else {
						high = middle - 1 // select left
					}
				}

				// last item slice length 1
				if high == low && toc_names.TOCList[low].Toc == cell_data {
					// found at last
					dialog.ShowInformation("TOC Name", toc_names.TOCList[low].Name, mywin_obj)
					return
				} else {
					// not found
					dialog.ShowError(errors.New("invalid TOC name"), mywin_obj)
					return
				}

			}
		}
	}
	return table
}

type settings struct {
	Freq        float64 `json:"freq"`
	Key         string  `json:"key"`
	Desired_len int     `json:"desired_len"`
}

func crs_to_name(crs string) (string, error) {
	if crs == "*" {
		return "Any Station", nil
	}

	// binary search for station
	// O(log n) instead of O(n)
	// the sorted list has almost 3000 items
	// 12 times max
	low := 0
	high := len(all_stations.StationList)
	for high > low {
		middle := (low + high) / 2 // floor

		if all_stations.StationList[middle].Crs == crs {
			return all_stations.StationList[middle].Name, nil
		} else if crs > all_stations.StationList[middle].Crs {
			low = middle + 1 // select right
		} else {
			high = middle - 1 // select left
		}
	}

	// last item slice length 1
	if high == low && all_stations.StationList[low].Crs == crs {
		// found at last
		return all_stations.StationList[low].Name, nil
	} else {
		return "Unknown Station", errors.New("unknown station") // not found
	}
}

// use configured data to get data of train services
func trains(key string, rootURI fyne.URI, numRows int) ([][]train_service, [][2]string, int, error) {
	const base_url string = "https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepartureBoard/"

	//crs = strings.ToUpper(strings.TrimSpace(crs))

	// var dep_api_key = os.Getenv("dep_key")
	_, qts, err := load_json("qtt.json", rootURI)
	if err != nil {
		return nil, nil, 0, err
	}

	var now time.Time
	const current_tz string = " UTC"
	utcNow := time.Now().UTC()

	if IsUKUsingSummerTime() {
		now = utcNow.Add(time.Hour)
	}

	var today int = int(now.Weekday())
	correct_time := make([]quick_time, 0)

	date_only := now.Format(time.RFC822)
	date_only = date_only[0:10]
	var correct_count int

	for _, qt := range qts.Quick_times {
		// if today is chosen
		if slices.Contains(qt.Days, today) {
			start, err := time.Parse(time.RFC822, date_only+qt.Start+current_tz)
			if err != nil {
				return nil, nil, 0, err
			}
			end, err := time.Parse(time.RFC822, date_only+qt.End+current_tz)
			if err != nil {
				return nil, nil, 0, err
			}

			// if within time range
			if now.After(start) && now.Before(end) {
				correct_time = append(correct_time, qt)
				correct_count++
				if correct_count >= 2 {
					break // more than two within time
				}
			}
		}
	}
	if len(correct_time) == 0 {
		return nil, nil, 0, nil // not in any time ranges
	}
	res := make([][]train_service, 0, len(correct_time))
	f_t_list := make([][2]string, 0, len(correct_time))
	for _, v := range correct_time {
		var url string = base_url + v.Org
		var params string
		var err error
		if v.Dest != "*" {
			params, err = format_params([]string{"filterCrs", "filterType", "numRows"},
				[]string{v.Dest, "to", fmt.Sprint(numRows)}) // has filter dest
		} else {
			params, err = format_params([]string{"numRows"},
				[]string{fmt.Sprint(numRows)}) // no filter dest
		}

		if err != nil {
			return nil, nil, 0, err
		} else {
			url = url + params // concat paremeters to url
		}
		this_res, err := request(url, key)
		if err != nil {
			return nil, nil, 0, err
		}
		res = append(res, this_res) // append this request to list of requests

		org_name, err := crs_to_name(v.Org)
		if err != nil {
			return nil, nil, 0, err
		}
		dest_name, err := crs_to_name(v.Dest)
		if err != nil {
			return nil, nil, 0, err
		}

		f_t_list = append(f_t_list, [2]string{fmt.Sprintf("%s to %s", v.Org, v.Dest),
			fmt.Sprintf("%s to %s", org_name, dest_name)})
		// title string list
	}
	return res, f_t_list, correct_count, nil
}

func apply_col_widths(table *widget.Table) {
	table.SetColumnWidth(-1, 30) // number header
	table.SetColumnWidth(0, 40)  // plat
	table.SetColumnWidth(1, 40)  // toc
	table.SetColumnWidth(2, 60)  // std
	table.SetColumnWidth(3, 50)  // dest
	table.SetColumnWidth(4, 80)  // etd
}

func tt_table(ut []train_service, dl int, ch, rh []string, mywin_addr *fyne.Window) *widget.Table {
	var data [][]string
	var datarow []string
	for i, val := range ut {
		if i >= dl {
			break // if too many services
		}
		datarow = nil
		datarow = append(datarow, val.plat, val.toc, val.std, val.dest, val.etd)
		data = append(data, datarow)
	}

	config := NewTableConfig(data, ch, rh)
	config.CellTemplateText = "?" // put ? for unknown data
	table := config.BuildTable(mywin_addr)
	apply_col_widths(table)
	return table
}

// update main label (get data + gui)
func refershTimes(mylabel_addr **widget.Label,
	mywin_addr *fyne.Window,
	hometab_addr **container.TabItem,
	apptabs_addr **container.AppTabs,
	key string,
	desired_len int,
	rootURI fyne.URI,
	ref_button **widget.Button) {
	apptabs_obj := *apptabs_addr
	if apptabs_obj.SelectedIndex() != 0 {
		return // not on this page
	}

	mylabel_obj := *mylabel_addr
	ref_button_obj := *ref_button
	fyne.Do(func() {
		mylabel_obj.SetText("refreshing train times") // refresh message
	})

	mywin_obj := *mywin_addr

	updated_times_s, f_t_list, correct_count, err := trains(key, rootURI, desired_len)
	if err != nil {
		dialog.ShowError(err, mywin_obj)
	}

	hometab_obj := *hometab_addr

	colHeaders := []string{"Plat", "TOC", "STD", "Dest", "ETD"}
	var rowHeaders []string
	for i := range desired_len {
		rowHeaders = append(rowHeaders, fmt.Sprintf("%v", i+1))
	}
	switch correct_count {
	case 0: // no correct
		mylabel_obj := *mylabel_addr
		fyne.Do(func() {
			mylabel_obj.SetText("not in specified time frames")
			hometab_obj.Content = container.NewBorder(container.NewHBox(ref_button_obj, mylabel_obj), nil, nil, nil, nil)
		})
	case 1: // one correct, whole page
		table := tt_table(updated_times_s[0], desired_len, colHeaders, rowHeaders, mywin_addr)
		fyne.Do(func() {
			mylabel_obj.SetText("")
			hometab_obj.Content = container.NewBorder(container.NewHBox(ref_button_obj, mylabel_obj), nil, nil, nil, container.NewScroll(widget.NewCard(f_t_list[0][0], f_t_list[0][1], table)))
		})

	case 2: // two correct, split page
		table := tt_table(updated_times_s[0], desired_len, colHeaders, rowHeaders, mywin_addr)
		table2 := tt_table(updated_times_s[1], desired_len, colHeaders, rowHeaders, mywin_addr)

		fyne.Do(func() {
			mylabel_obj.SetText("")

			hometab_obj.Content = container.NewBorder(container.NewHBox(ref_button_obj, mylabel_obj), nil, nil, nil,
				container.New(NewHalfHeightLayout(),
					container.NewScroll(
						widget.NewCard(f_t_list[0][0], f_t_list[0][1], table)),
					container.NewScroll(
						widget.NewCard(f_t_list[1][0], f_t_list[1][1], table2)),
				))

		})
	default:
		dialog.ShowConfirm("something went wrong", fmt.Sprintf("incorrect number of correct times (%v)", correct_count), nil, mywin_obj)
	} // function shouldn't return more than two but just in case

}

func tidyUp() {
	// fmt.Println("exiting app, thank you for using Quick Train Times")
}

// ----- global vars -----
var toc_names tocs

// ##### main #####
func main() {
	// run when exiting
	defer tidyUp()

	myapp := app.NewWithID("qtt")
	mywin := myapp.NewWindow("Quick Train Times")
	mywin.Resize(fyne.NewSize(640, 640))

	placeholder := widget.NewLabel("train times go here")
	refresh_button := widget.NewButton("refresh manually", func() {})
	refresh_button.Show()
	top_bar := container.NewHBox(refresh_button, placeholder)
	home_border := container.NewBorder(top_bar, nil, nil, nil, nil)
	home_tab := container.NewTabItem("Home", home_border)

	rootURI := myapp.Storage().RootURI()

	ver, err := get_ver()
	if err != nil {
		dialog.ShowError(err, mywin)
	} else {
		mywin.SetTitle("Quick Train Times " + ver)
	}

	// -----settings page------
	entry_freq := widget.NewEntry()
	entry_freq.SetPlaceHolder("in seconds")

	entry_freq.Validator = func(s string) error {
		num, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return errors.New("not a number")
		} else if num <= 0 {
			return errors.New("number must be positive")
		} else {
			return nil
		}
	}

	entry_key := widget.NewEntry()
	entry_key.SetPlaceHolder("48 character long key")

	entry_key.Validator = func(s string) error {
		if len(s) != 48 {
			return errors.New("invalid key")
		} else {
			return nil
		}
	}

	entry_len := widget.NewEntry()
	entry_len.SetPlaceHolder("positive integer [1,150]")
	entry_len.Validator = func(s string) error {
		myint, err := strconv.Atoi(s)
		if err != nil {
			return errors.New("not an integer")
		} else if myint <= 0 || myint > 150 {
			return errors.New("not within [1,150]")
		} else {
			return nil
		}

	}

	existing_settings, _, err := load_json("settings.json", rootURI)
	if err != nil {
		dialog.ShowError(err, mywin)
	}

	entry_freq.SetText(fmt.Sprint(existing_settings.Freq))
	entry_key.SetText(existing_settings.Key)
	entry_len.SetText(fmt.Sprint(existing_settings.Desired_len))

	form := &widget.Form{
		OnSubmit: func() { // optional, handle form submission
			var s settings
			s.Freq, err = strconv.ParseFloat(entry_freq.Text, 64)
			s.Key = entry_key.Text
			s.Desired_len, _ = strconv.Atoi(entry_len.Text)
			err := save_json(s, "settings.json", rootURI)
			if err != nil {
				dialog.ShowError(err, mywin)
			} else {
				dialog.ShowInformation("Info", "Settings saved successfully. Not effective until program restart.", mywin)
			}
		},
		OnCancel: func() {
			entry_freq.SetText(fmt.Sprint(existing_settings.Freq))
			entry_key.SetText(existing_settings.Key)
			entry_len.SetText(fmt.Sprint(existing_settings.Desired_len))
		},
	}

	// append items to form
	form.Append("Refresh Frequency (secs)", entry_freq)
	form.Append("Departure API Key", entry_key)
	form.Append("Max num of train times", entry_len)
	form.SubmitText = "Save"

	con := container.NewBorder(nil, widget.NewLabel("Changes will be applied next time starting the program."), nil, nil, form)
	settings_tab := container.NewTabItem("Settings", con)

	config_tab := container.NewTabItem("Config QTTs", qtt_init(&mywin, rootURI))

	mytabs := container.NewAppTabs(home_tab, settings_tab, config_tab)
	mywin.SetContent(mytabs)
	refershTimes(&placeholder, &mywin, &home_tab, &mytabs, existing_settings.Key, existing_settings.Desired_len, rootURI, &refresh_button)

	refresh_button.OnTapped = func() {
		fyne.Do(func() { placeholder.SetText("refreshing train times") })
		go refershTimes(&placeholder, &mywin, &home_tab, &mytabs, existing_settings.Key, existing_settings.Desired_len, rootURI, &refresh_button)
		fyne.Do(func() { mywin.SetContent(mytabs) })
	}

	// when going to main tab refresh train time
	mytabs.OnSelected = func(selectedTab *container.TabItem) {
		if mytabs.SelectedIndex() == 0 {
			fyne.Do(func() { placeholder.SetText("refreshing train times") })
			go refershTimes(&placeholder, &mywin, &home_tab, &mytabs, existing_settings.Key, existing_settings.Desired_len, rootURI, &refresh_button)
			fyne.Do(func() { mywin.SetContent(mytabs) })
		} else {
			placeholder.SetText("refreshing train times")
			// home_tab.Content = placeholder
		}
	}

	// load TOC names to global var
	err = json.Unmarshal(resourceTocJson.StaticContent, &toc_names)
	if err != nil {
		dialog.ShowError(err, mywin)
	}

	go func() {
		// main loop
		for range time.Tick(time.Second * time.Duration(existing_settings.Freq)) {
			refershTimes(&placeholder, &mywin, &home_tab, &mytabs, existing_settings.Key, existing_settings.Desired_len, rootURI, &refresh_button)
			fyne.Do(func() { mywin.SetContent(mytabs) })
		}
	}()

	mywin.Show()
	myapp.Run()

}
