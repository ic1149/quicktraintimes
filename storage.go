package main

import (
	"encoding/json"
	"io"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/storage"
	"github.com/pelletier/go-toml/v2"
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

func load_json(fname string, rootURI fyne.URI) (settings, qtt, error) {
	default_strs := map[string]string{
		"settings.json": `{"freq":60,"key":"ZF2QTLjOalPE2KrbeoUsOarJ7ic4XQHJPnR9eiSHR9I4j0A0","desired_len":5}`,
		"qtt.json":      `{"quick_times":[{}]}`,
	}
	mysettings := settings{Freq: 60, Key: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"}
	myqtt := qtt{Quick_times: make([]quick_time, 0), del_ids: make([]int, 0)}

	myURI, err := storage.Child(rootURI, fname)
	if err != nil {
		return mysettings, myqtt, err
	}

	exists, err := storage.Exists(myURI)
	if err != nil {
		return mysettings, myqtt, err
	}

	if !exists {
		// If the file does not exist, create it and write the default content
		writeCloser, err := storage.Writer(myURI)
		if err != nil {
			return mysettings, myqtt, err
		}
		defer writeCloser.Close()

		_, err = io.WriteString(writeCloser, default_strs[fname])
		if err != nil {
			return mysettings, myqtt, err
		}
		return mysettings, myqtt, nil
	} else {
		readCloser, err := storage.Reader(myURI)
		if err != nil {
			return mysettings, myqtt, err
		}
		defer readCloser.Close()

		content, err := io.ReadAll(readCloser)
		if err != nil {
			return mysettings, myqtt, err
		}
		switch fname {
		case "settings.json":
			err = json.Unmarshal(content, &mysettings)
		case "qtt.json":
			err = json.Unmarshal(content, &myqtt)
		}
		if err != nil {
			return mysettings, myqtt, err
		}
		return mysettings, myqtt, nil
	}
}

func get_ver() (string, error) {
	var met metadata
	err := toml.Unmarshal(resourceFyneAppToml.StaticContent, &met)
	if err != nil {
		return "", err
	}
	return met.Details.Version, nil

}
