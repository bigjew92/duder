package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func getRugStorageFile(rug Rug) string {
	return strings.TrimSuffix(rug.file, filepath.Ext(rug.file)) + ".json"
}

// LoadRugStorage description
func LoadRugStorage(rug Rug) (string, error) {
	//path := strings.TrimSuffix(rug.File, filepath.Ext(rug.File))
	path := getRugStorageFile(rug)

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Storage file for '%v' not found; creating new one...", rug.name)

		// create the storage file
		if e := ioutil.WriteFile(path, []byte("{}"), 0644); e != nil {
			return "{}", errors.New(fmt.Sprint("unable to create storage file ", path, e.Error()))
		}
		log.Printf("Storage file %v created", path)
	} else {
		var bytes []byte
		if bytes, err = ioutil.ReadFile(path); err != nil {
			return "{}", errors.New(fmt.Sprint("unable to read storage file ", path, err.Error()))
		}

		return string(bytes), nil
	}
	return "{}", nil
}
