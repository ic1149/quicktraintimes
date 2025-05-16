package main

import (
	"encoding/json"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
)

func save_json(mystruct any, fname string, rootURI fyne.URI) error {
	jsonData, err := json.Marshal(mystruct)
	if err != nil {
		return err
	}

	myURI, err := storage.Child(rootURI, fname)
	if err != nil {
		return err
	}

	writeCloser, err := storage.Writer(myURI)
	if err != nil {
		return err
	}
	defer writeCloser.Close()

	_, err = io.Writer.Write(writeCloser, jsonData)
	if err != nil {
		return err
	}

	return nil
}

func load_json(fname string, rootURI fyne.URI) (settings, error) {
	var mysettings settings
	mysettings.Freq = 60
	mysettings.Key = "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
	myURI, err := storage.Child(rootURI, fname)
	if err != nil {
		return mysettings, err
	}

	exists, err := storage.Exists(myURI)
	if err != nil {
		return mysettings, err
	}

	if !exists {
		// If the file does not exist, create it and write the default content
		defaultStr := `{
    "freq":60,
    "key":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
}` // Your default JSON string

		writeCloser, err := storage.Writer(myURI)
		if err != nil {
			return mysettings, err
		}
		defer writeCloser.Close()

		_, err = io.WriteString(writeCloser, defaultStr)
		if err != nil {
			return mysettings, err
		}
		return mysettings, nil
	} else {

		// Example of reading the existing file:
		readCloser, err := storage.Reader(myURI)
		if err != nil {
			return mysettings, err
		}
		defer readCloser.Close()

		content, err := io.ReadAll(readCloser)
		if err != nil {
			return mysettings, err
		}
		err = json.Unmarshal(content, &mysettings)
		if err != nil {
			return mysettings, err
		}
		return mysettings, nil
	}
}
