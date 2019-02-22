package stash

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"gittest.impetus.co.in/Codeathon/securevault/stashapp/utils"
)

const (
	//TempResourcePath ...
	TempResourcePath = "/tmp/7y88u9u"
)

//TempJSON ...
type TempJSON struct {
	Data []byte `json:"Data"`
	Name string `json:"Name"`
	Hash []byte `json:"Hash"`
}

//CommandImpl implemenation
type CommandImpl struct {
	kvmap        map[string][]byte
	wallet       []byte
	InteralStore Store
	Secret       []byte
}

/*Create Creates data from a source and places it in temp location after encoding it*/
func (s *CommandImpl) Create(name string, hash []byte) error {
	var kvmap map[string]json.RawMessage

	//read from store
	raw, err := s.InteralStore.Read(name)
	if err != nil {
		return errors.New("[Create failed] Failed to read from remote location: " + err.Error())
	}
	//Dump into kvmap obj
	err = json.Unmarshal(raw, &kvmap)
	if err != nil {
		return errors.New("[Create failed] Data not in correct format: " + err.Error())
	}
	s.kvmap = make(map[string][]byte)
	for k, v := range kvmap {
		s.kvmap[k] = []byte(v)
	}

	//encrypt
	err = s.Encrypt()
	if err != nil {
		return errors.New("[Create failed]" + err.Error())
	}

	//Place it in tmp
	hash, err = utils.CreateHashPassphrase(hash)
	if err != nil {
		return errors.New("[Create failed]" + err.Error())
	}
	tj := TempJSON{Data: []byte(s.wallet), Name: name, Hash: hash}
	err = WriteJSONToTmp(tj)
	return err
}

//Fetch ...
func (s *CommandImpl) Fetch(name string) error {
	raw, err := s.InteralStore.Read(name)
	if err != nil {
		return errors.New("[Fetch failed] Could not read from the remote datasource. " + err.Error())
	}
	var tj TempJSON
	err = json.Unmarshal(raw, &tj)
	if err != nil {
		return errors.New("[Fetch failed] Could not unmarshal data from the remote datasource. " + err.Error())
	}
	err = WriteJSONToTmp(tj)
	if err != nil {
		return errors.New("[Fetch failed] Could not write to tmp location from the remote datasource. " + err.Error())
	}
	return err
}

//Publish ...
func (s *CommandImpl) Publish(name string) error {
	tj, err := ReadTmpToJSON()
	if err != nil {
		return errors.New("[Publish failed] Could not read from tmp location " + err.Error())
	}
	//now temporary and remote point to this new name
	tj.Name = name
	WriteJSONToTmp(tj)
	//err := s.InteralStore.Write(name, []byte(tj.Data))
	data, err := json.Marshal(tj)
	if err != nil {
		return errors.New("[Publish failed] Could not marshal data " + err.Error())
	}
	err = s.InteralStore.Write(name, data)
	if err != nil {
		return errors.New("[Publish failed] Could not write to remote location " + err.Error())
	}
	return err
}

//Encrypt ...
func (s *CommandImpl) Encrypt() error {
	kvmap := make(map[string][]byte)
	//Encrypt values
	for k, v := range s.kvmap {
		//binary encoding data

		encodedData, err := encryptData(v, s.Secret)
		if err != nil {
			fmt.Println("[WARN] Skipping encoding value of key: "+k, err)
		}
		kvmap[k] = []byte(encodedData)
	}
	//Encrypt root
	//NOTE: JSON marshalling in go automatically base64 encodes a byte array
	str, err := json.Marshal(kvmap)
	if err != nil {
		return errors.New("[Encryption failed] Failed to unmarshal encrypted data " + err.Error())
	}
	encyptedData, err := encryptData(str, s.Secret)
	s.wallet = encyptedData
	if err != nil {
		return errors.New("[Encryption failed] Failed to encrypt the complete map " + err.Error())
	}
	return nil
}

//Decrypt ...
func (s *CommandImpl) Decrypt() (map[string][]byte, error) {
	var ret map[string][]byte
	//Decrypt root
	var decodedString []byte
	decodedString, err := decryptData([]byte(s.wallet), s.Secret)
	if err != nil {
		return ret, errors.New("[Decryption failed] Failed to decrypt the root data " + err.Error())
	}
	err = json.Unmarshal(decodedString, &ret)
	if err != nil {
		return ret, errors.New("[Decryption failed] Failed to unmarshal root data" + err.Error())
	}

	//decrypt values
	for k, v := range ret {
		decodedValue, err := decryptData(v, s.Secret)
		if err != nil {
			fmt.Println("[WARN] Skipping decrypting value of key: "+k, err)
			continue
		}

		ret[k] = decodedValue
	}

	return ret, nil
}

//Pull Pulls updates from local copy
func (s *CommandImpl) Pull() error {
	tj, err := ReadTmpToJSON()
	if err != nil {
		return errors.New("[Pull failed]" + err.Error())
	}

	s.wallet = tj.Data
	//decrypts the wallet and sets
	kvmap, err := s.Decrypt()
	if err != nil {
		return errors.New("[Pull failed]" + err.Error())
	}
	s.kvmap = kvmap
	return nil
}

//Push Pushed updates to local copy
func (s *CommandImpl) Push() error {
	err := s.Encrypt()
	if err != nil {
		return errors.New("[Push failed]" + err.Error())
	}

	err = updateTemp(s.wallet)
	if err != nil {
		return errors.New("[Pull failed]" + err.Error())
	}

	return nil
}

//Get ..
func (s *CommandImpl) Get(pattern string) ([]string, error) {
	var ret []string
	if val, ok := s.kvmap[pattern]; ok {
		ret = append(ret, string(val))
		return ret, nil
	}

	regex, err := regexp.Compile(pattern)
	if err != nil {
		return ret, errors.New("[Get failed] Illegal pattern: " + pattern)
	}

	for k, v := range s.kvmap {
		if regex.MatchString(strings.ToLower(k)) {
			//attack the key to output too. In json format so we can prettify it later
			out := "{\"\u0F3A " + k + "\u0F3B  \":" + string(v) + "},"
			ret = append(ret, out)
		}
	}
	if len(ret) < 1 {
		err = errors.New("[Get failed] No key found")
		return ret, err
	}
	//remove the last comma from the last element in the array
	ret[len(ret)-1] = strings.TrimSuffix(ret[len(ret)-1], ",")

	return ret, err
}

//Delete ..
func (s *CommandImpl) Delete(key string) error {
	delete(s.kvmap, key)
	//update wallet
	return s.Encrypt()
}

//AddOrUpdate ...
func (s *CommandImpl) AddOrUpdate(key string, value []byte) error {
	s.kvmap[key] = value
	return s.Encrypt()
}

//GetHash ...
func (s *CommandImpl) GetHash() ([]byte, error) {
	tj, err := ReadTmpToJSON()
	if err != nil {
		return nil, errors.New("[Crypto] Failed to fetch local hash")
	}
	return []byte(tj.Hash), nil
}

//internals
func encryptData(data []byte, key []byte) ([]byte, error) {
	//return base64.StdEncoding.EncodeToString(data), nil
	return utils.Encrypt(data, key), nil
}
func decryptData(data []byte, key []byte) ([]byte, error) {
	//return base64.StdEncoding.DecodeString(data)
	return utils.Decrypt(data, key), nil
}

//ReadTmpToJSON ...
func ReadTmpToJSON() (TempJSON, error) {
	var f TempJSON
	raw, err := ioutil.ReadFile(TempResourcePath)
	if err != nil {
		return f, errors.New("Can't read from temp file. " + err.Error())
	}
	err = json.Unmarshal(raw, &f)
	if err != nil {
		return f, errors.New("Can't read from temp file. " + err.Error())
	}
	return f, nil
}

//WriteJSONToTmp ...
func WriteJSONToTmp(f TempJSON) error {
	raw, err := json.Marshal(f)
	if err != nil {
		return errors.New("Can't write to temp file. " + err.Error())
	}
	err = ioutil.WriteFile(TempResourcePath, raw, 0664)
	if err != nil {
		return errors.New("Can't write to temp file. " + err.Error())
	}
	return nil
}

func updateTemp(data []byte) error {
	tj, err := ReadTmpToJSON()
	if err != nil {
		return errors.New("Can't read from temp file. " + err.Error())
	}
	tj.Data = data
	err = WriteJSONToTmp(tj)
	return err
}
