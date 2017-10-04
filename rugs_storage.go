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

// LoadRugStorage description
func LoadRugStorage(rug Rug) (string, error) {
	path := strings.TrimSuffix(rug.Path, filepath.Ext(rug.Path))

	// check if the file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Printf("Storage file for '%v' not found; creating new one...", rug.Name)

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
