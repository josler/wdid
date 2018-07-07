package auto

// much of this from https://developers.google.com/calendar/quickstart/go

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
)

type GoogleCalendar struct {
	calendarName string
}

func NewGoogleCalendar(calendarName string) *GoogleCalendar {
	return &GoogleCalendar{calendarName: calendarName}
}

func (gc *GoogleCalendar) Precheck() {
	gc.getClient(gc.getConfig())
}

func (gc *GoogleCalendar) Load(startTime, endTime time.Time) []*Option {
	srv, err := calendar.New(gc.getClient(gc.getConfig()))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	events, err := srv.Events.List(gc.calendarName).ShowDeleted(false).
		SingleEvents(true).
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		MaxResults(10).
		OrderBy("startTime").
		Do()

	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events: %v", err)
	}

	options := []*Option{}

	for _, item := range events.Items {
		date := item.Start.DateTime
		if date == "" {
			date = item.Start.Date
		}
		parsedTime, err := time.Parse(time.RFC3339, date)
		if err != nil {
			continue
		}
		status := "waiting"
		if parsedTime.Before(time.Now()) {
			status = "done"
		}
		opt := Option{
			data:     fmt.Sprintf("[%s] %s %s", "meeting", item.Summary, item.HangoutLink),
			dateTime: parsedTime,
			status:   status,
		}
		options = append(options, &opt)
	}

	return options
}

// Retrieve a token, saves the token, then returns the generated client.
func (gc *GoogleCalendar) getClient(config *oauth2.Config) *http.Client {
	tokFile := "google_calendar_token.json"
	tok, err := gc.tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		gc.saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func (gc *GoogleCalendar) tokenFromFile(path string) (*oauth2.Token, error) {
	f, err := os.Open(filepath.Join(configDir(), path))
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func (gc *GoogleCalendar) saveToken(path string, token *oauth2.Token) {
	f, err := os.OpenFile(filepath.Join(configDir(), path), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func (gc *GoogleCalendar) getConfig() *oauth2.Config {
	b, err := ioutil.ReadFile(filepath.Join(configDir(), "client_secret.json"))
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	return config
}

func configDir() string {
	return filepath.Join(homeDir(), ".config", "wdid")
}

func homeDir() string {
	usr, err := user.Current()
	if err != nil {
		return "defaultuser"
	}
	return usr.HomeDir
}
