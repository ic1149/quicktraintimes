package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

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
	Start string `json:"start"`
	End   string `json:"end"`
	Org   string `json:"org"`
	Dest  string `json:"dest"`
	Days  []int  `json:"days"`
}

// catch qtt.json
type config struct {
	Quick_times []quick_time `json:"quick_times"`
}

type TableConfig struct {
	Data               [][]string // The actual data for the table cells
	ColHeaderTexts     []string   // Texts for the column headers
	RowHeaderTexts     []string   // Texts for the row headers
	CellTemplateText   string     // Placeholder text for data cell templates
	HeaderTemplateText string     // Placeholder text for header cell templates
	CornerHeaderText   string     // Text for the top-left corner header cell
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

func (tc *TableConfig) BuildTable() *widget.Table {
	// Basic validation for data
	if tc.Data == nil {
		log.Println("Warning: TableConfig.Data is nil. Creating an empty table.")
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

		// id.Row == -1 && id.Col == -1: Top-left corner header cell
		// id.Row == -1 && id.Col >= 0: Column header cell
		// id.Col == -1 && id.Row >= 0: Row header cell

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

	return table
}

type settings struct {
	Freq float64 `json:"freq"`
	Key  string  `json:"key"`
}

func load_settings() settings {
	var s settings
	b, err := os.ReadFile("settings.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(b, &s)
	return s
}

func load_key() string {
	var s settings = load_settings()
	return s.Key
}

// load data from qtt.json
func load_qtt() []quick_time {
	b, err := os.ReadFile("qtt.json")
	if err != nil {
		fmt.Println(err)
	}
	var c config
	err = json.Unmarshal(b, &c)
	if err != nil {
		fmt.Println(err)
	}

	// quick_times := make([]quick_time, 0, len(c.Quick_times))
	// for _, val := range c.Quick_times {
	// 	this_map := val.(map[string]any)
	// 	var this_struct quick_time
	// 	this_struct.start = this_map["start"].(string)
	// 	this_struct.end = this_map["end"].(string)
	// 	this_struct.org = this_map["org"].(string)
	// 	this_struct.dest = this_map["dest"].(string)
	// 	for _, val := range this_map["days"].([]any) {
	// 		this_struct.days = append(this_struct.days, int(val.(float64)))
	// 	}

	// 	quick_times = append(quick_times, this_struct)
	// }
	return c.Quick_times
}

// use configured data to get data of train services
func trains() ([][]train_service, []string, int) {
	// err := godotenv.Load(".env")
	// if err != nil {
	// 	log.Fatalf("Error loading .env file: %s", err)
	// }

	const base_url string = "https://api1.raildata.org.uk/1010-live-departure-board-dep1_2/LDBWS/api/20220120/GetDepartureBoard/"

	//crs = strings.ToUpper(strings.TrimSpace(crs))

	// var dep_api_key = os.Getenv("dep_key")
	dep_api_key, qts := load_key(), load_qtt()
	now := time.Now()
	var today int = int(now.Weekday())
	correct_time := make([]quick_time, 0)
	current_tz, _ := now.Zone()
	current_tz = " " + current_tz
	date_only := now.Format(time.RFC822)
	date_only = date_only[0:10]
	var correct_count int
	for _, qt := range qts {
		if slices.Contains(qt.Days, today) {
			start, err := time.Parse(time.RFC822, date_only+qt.Start+current_tz)
			if err != nil {
				fmt.Println(err)
			}
			end, err := time.Parse(time.RFC822, date_only+qt.End+current_tz)
			if err != nil {
				fmt.Println(err)
			}

			if now.After(start) && now.Before(end) {
				correct_time = append(correct_time, qt)
				correct_count++
				if correct_count >= 2 {
					break
				}
			}
		}
	}
	if len(correct_time) == 0 {
		return nil, nil, 0
	}
	res := make([][]train_service, 0, len(correct_time))
	f_t_list := make([]string, 0, len(correct_time))
	for _, v := range correct_time {
		params, err := format_params([]string{"filterCrs", "filterType"},
			[]string{v.Dest, "to"})

		if err != nil {
			fmt.Println(err)
		}
		var url string = base_url + v.Org + params
		res = append(res, request(url, dep_api_key))
		f_t_list = append(f_t_list, fmt.Sprintf("%s to %s", v.Org, v.Dest))
	}
	return res, f_t_list, correct_count
}

// update main label (get data + gui)
func refershTimes(mylabel_addr **widget.Label, mywin_addr *fyne.Window, hometab_addr **container.TabItem, apptabs_addr **container.AppTabs) {
	apptabs_obj := *apptabs_addr
	if apptabs_obj.SelectedIndex() != 0 {
		return
	}

	fmt.Println("refreshing train times")
	const desired_len = 5
	updated_times_s, f_t_list, correct_count := trains()

	mywin_obj := *mywin_addr
	hometab_obj := *hometab_addr

	colHeaders := []string{"Plat", "TOC", "STD", "Dest", "ETD"}
	var rowHeaders []string
	for i := range desired_len {
		rowHeaders = append(rowHeaders, fmt.Sprintf("%v", i+1))
	}
	switch correct_count {
	case 0:
		mylabel_obj := *mylabel_addr
		fyne.Do(func() {
			mylabel_obj.SetText("not in specified time frames")
			hometab_obj.Content = mylabel_obj
		})
	case 1:
		var data [][]string
		var datarow []string
		for _, val := range updated_times_s[0] {
			datarow = nil
			datarow = append(datarow, val.plat, val.toc, val.std, val.dest, val.etd)
			data = append(data, datarow)
		}

		config := NewTableConfig(data, colHeaders, rowHeaders)
		config.CellTemplateText = "?"
		table := config.BuildTable()
		fyne.Do(func() {
			hometab_obj.Content = container.NewScroll(widget.NewCard(f_t_list[0], "", table))
		})

	case 2:
		var data [][]string
		var datarow []string
		for _, val := range updated_times_s[0] {
			datarow = nil
			datarow = append(datarow, val.plat, val.toc, val.std, val.dest, val.etd)
			data = append(data, datarow)
		}

		config := NewTableConfig(data, colHeaders, rowHeaders)
		config.CellTemplateText = "N/A"
		table := config.BuildTable()

		var data2 [][]string
		var datarow2 []string
		for _, val := range updated_times_s[1] {
			datarow2 = nil
			datarow2 = append(datarow2, val.plat, val.toc, val.std, val.dest, val.etd)
			data2 = append(data2, datarow2)
		}

		config2 := NewTableConfig(data2, colHeaders, rowHeaders)
		config2.CellTemplateText = "N/A"
		table2 := config2.BuildTable()
		fyne.Do(func() {
			hometab_obj.Content = container.NewVSplit(container.NewScroll(widget.NewCard(f_t_list[0], "", table)), container.NewScroll(widget.NewCard(f_t_list[1], "", table2)))
		})
	default:
		dialog.ShowConfirm("something went wrong", fmt.Sprintf("incorrect number of correct times (%v)", correct_count), nil, mywin_obj)
	}

}

func tidyUp() {
	fmt.Println("exiting app, thank you for using Quick Train Times")
}

func main() {
	// run when exiting
	defer tidyUp()

	myapp := app.New()
	mywin := myapp.NewWindow("Quick Train Times")

	placeholder := widget.NewLabel("train times go here")
	home_tab := container.NewTabItem("Home", placeholder)
	settings_tab := container.NewTabItem("Settings", widget.NewLabel("settings will be here"))
	config_tab := container.NewTabItem("Config QTTs", widget.NewLabel("config quick train times"))
	mytabs := container.NewAppTabs(home_tab, settings_tab, config_tab)
	mywin.SetContent(mytabs)
	refershTimes(&placeholder, &mywin, &home_tab, &mytabs)
	mywin.SetContent(mytabs)

	mytabs.OnSelected = func(selectedTab *container.TabItem) {
		if mytabs.SelectedIndex() == 0 {
			go refershTimes(&placeholder, &mywin, &home_tab, &mytabs)
			fyne.Do(func() { mywin.SetContent(mytabs) })
		} else {
			placeholder.SetText("getting train times")
			home_tab.Content = placeholder
		}
	}

	go func() {
		// every minute
		for range time.Tick(time.Second * 10) {
			refershTimes(&placeholder, &mywin, &home_tab, &mytabs)
			fyne.Do(func() { mywin.SetContent(mytabs) })
		}
	}()

	mywin.Resize(fyne.NewSize(640, 640))
	mywin.Show()
	myapp.Run()

}
