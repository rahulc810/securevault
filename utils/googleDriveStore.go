package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

const (
	clientID = `{
   "installed":{
      "client_id":"689461976475-segtkobedosa433dlpgl3teu9nggk99e.apps.googleusercontent.com",
      "project_id":"decent-carving-188410",
      "auth_uri":"https://accounts.google.com/o/oauth2/auth",
      "token_uri":"https://accounts.google.com/o/oauth2/token",
      "auth_provider_x509_cert_url":"https://www.googleapis.com/oauth2/v1/certs",
      "client_secret":"qNw70b2-QZs2U6ilyF8RUMy-",
      "redirect_uris":[
         "urn:ietf:wg:oauth:2.0:oob",
         "http://localhost"
      ]
   }
	}
`
)

//GooleDriveStore ...
type GooleDriveStore struct {
	srv *drive.Service
}

//NewGoogleStore ...
func NewGoogleStore() GooleDriveStore {
	ctx := context.Background()
	// b, err := ioutil.ReadFile("client_secret.json")
	// if err != nil {
	// 	log.Fatalf("Unable to read client secret file: %v", err)
	// }

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/drive-go-quickstart.json
	config, err := google.ConfigFromJSON([]byte(clientID), drive.DriveScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := drive.New(client)

	if err != nil {
		log.Fatalf("Unable to retrieve drive Client %v", err)
		panic(err)
	}

	return GooleDriveStore{srv: srv}
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("drive-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

//DownloadFile ...
func DownloadFile(d *drive.Service, t http.RoundTripper, f *drive.File) (string, error) {
	// t parameter should use an oauth.Transport
	downloadUrl := f.WebContentLink
	if downloadUrl == "" {
		// If there is no downloadUrl, there is no body
		fmt.Printf("An error occurred: File is not downloadable")
		return "", nil
	}
	req, err := http.NewRequest("GET", downloadUrl, nil)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}
	resp, err := t.RoundTrip(req)
	// Make sure we close the Body later
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return "", err
	}
	return string(body), nil
}

//GetData ...
func (g GooleDriveStore) getData(name string) (*drive.File, error) {
	r, err := g.srv.Files.List().Q("name= '" + name + "'").Do()

	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
	}

	fmt.Println("Uploading files to google drive. Please wait...")
	if len(r.Files) > 0 {
		return r.Files[0], nil
	}
	return nil, errors.New("No files found")
}

func (g GooleDriveStore) Read(name string) ([]byte, error) {
	file, err := g.getData(name)
	r, err := g.srv.Files.Get(file.Id).Download()
	if err != nil {
		return []byte{}, errors.New("Could not download data from Drive. " + err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return []byte{}, errors.New("Could not download data from Drive. Empty response " + err.Error())
	}
	return body, nil
}

func (g GooleDriveStore) Write(name string, input []byte) error {
	file, err := g.getData(name)

	reader := bytes.NewReader(input)

	if err != nil {
		//File doesnt exist create
		newFile := &drive.File{Name: name}
		req := g.srv.Files.Create(newFile)
		_, err = req.Media(reader).Do()
	} else {
		newFile := &drive.File{Name: file.Name}
		req := g.srv.Files.Update(file.Id, newFile)
		_, err = req.Media(reader).Do()
	}

	if err != nil {
		return err
	}
	return nil
}
