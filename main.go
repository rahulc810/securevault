package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"gittest.impetus.co.in/Codeathon/securevault/stashapp/stash"
	"gittest.impetus.co.in/Codeathon/securevault/stashapp/utils"
)

type exec struct {
	operation  string
	key        string
	value      []byte
	connector  string
	name       string
	s          stash.Command
	passphrase []byte
}

func main() {
	ex := exec{}

	ex.operation = os.Args[1]
	arguments := os.Args[2:]
	ex.prepareArgMap(arguments)

	var connector stash.Store
	switch ex.connector {
	case "google":
		connector = utils.NewGoogleStore()
	default:
		connector = utils.FileStore{}
	}
	ex.s = &stash.CommandImpl{InteralStore: connector, Secret: ex.passphrase}

	if ok, msg := ex.validate(); !ok {
		fmt.Println(msg)
		return
	}

	var result, er interface{}
	switch ex.operation {
	case "create":
		er = ex.s.Create(ex.name, ex.passphrase)
	case "publish":
		if ex.name == "" {
			tj, _ := stash.ReadTmpToJSON()
			ex.name = tj.Name
		}
		er = ex.s.Publish(ex.name)
	case "fetch":
		if ex.name == "" {
			tj, _ := stash.ReadTmpToJSON()
			ex.name = tj.Name
		}
		er = ex.s.Fetch(ex.name)
	//Push and pull are useless for now. There is no concept of session
	// case "pull":
	// 	er = ex.s.Pull()
	// case "push":
	// 	er = ex.s.Push()
	case "get":
		er = ex.s.Pull()
		result, er = ex.s.Get(ex.key)
		if result != nil {
			var prettyJSON bytes.Buffer
			json.Indent(&prettyJSON, []byte(fmt.Sprintf("%v", result)), "", "  ")
			result = string(prettyJSON.Bytes())
		}
	case "add":
		er = ex.s.Pull()
		er = ex.s.AddOrUpdate(ex.key, ex.value)
		er = ex.s.Push()
	case "delete":
		ex.s.Pull()
		result = ex.s.Delete(ex.key)
		ex.s.Push()
	default:
		result = "Unrecognized operation"
	}

	if er != nil {
		fmt.Println("Failure to execute operation :", er)
	} else {
		if result != nil {
			fmt.Println(result)
		}
	}

}

func (e *exec) prepareArgMap(args []string) {
	for index, v := range args {
		if index%2 == 0 {
			switch v {
			//Operation
			case "-o":
				e.operation = args[index+1]
			//Key
			case "-k":
				e.key = args[index+1]
			//Value
			case "-v":
				e.value = []byte(args[index+1])
			//Connector
			case "-c":
				e.connector = args[index+1]
			//Name
			case "-n":
				e.name = args[index+1]
			case "-p":
				e.passphrase = []byte(args[index+1])
			}
		}

	}
}

func (e *exec) validate() (bool, string) {
	switch e.operation {
	case "create", "fetch":
		if e.passphrase == nil || len(e.passphrase) == 0 {
			return false, "[INFO]Invalid password. Please create a valid password."
		}
		return true, ""
	default:
		hash, err := e.s.GetHash()
		if err != nil {
			return false, "[INFO]Incorrect password===" + err.Error()
		}
		if !utils.VerifyPassphrase(e.passphrase, hash) {
			return false, "[INFO]Incorrect password.***"
		}
		return true, ""
	}
}
