package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

//FileStore implementation for store
type FileStore struct {
}

func (f FileStore) Read(name string) ([]byte, error) {
	raw, err := ioutil.ReadFile(name)
	if err != nil {
		fmt.Println("Could not read file: " + name)
		return []byte{}, errors.New("Could not read file: " + name + ". " + err.Error())
	}
	return raw, nil
}

func (f FileStore) Write(name string, input []byte) error {
	return ioutil.WriteFile(name, input, os.FileMode(0664))
}
