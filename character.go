package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
)

type Character struct {
	name     string
	endpoint string
	frames   map[string]FrameDatum
	logger   *log.Logger
}

type FrameDatum struct {
	s  string // startup
	a  string // active
	r  string // recovery
	ba string // block advantage
	ha string // hit advantage
}

func (character *Character) Log(msg string, vars ...string) {
	timestamp := Now()
	character.logger.Println(timestamp + " - " + msg + " " + strings.Join(vars, " "))
}

func newCharacter(name string, endpoint string, bot *Bot) *Character {
	character := &Character{
		name:     name,
		endpoint: endpoint,
		logger:   log.New(os.Stdout, "log: ", log.Lshortfile),
	}
	character.frames = SetFrames(character)
	return character
}

func SetFrames(character *Character) map[string]FrameDatum {
	uri := "http://localhost:8080/usf4/" + character.name

	character.Log("Fetching frames for " + character.name)
	resp, err := http.Get(uri)

	if err != nil {
		// handle error here
		character.Log("Failed to get frame data for " + character.name)
	}

	defer resp.Body.Close()
	body, httpReadErr := ioutil.ReadAll(resp.Body)

	if httpReadErr != nil {
		character.Log("Failed to read response body")
	}

	strValue := string(body[:])

	var rawJsonInterface interface{}
	jsonErr := json.Unmarshal(body, &rawJsonInterface)

	if jsonErr != nil {
		character.Log("Failed to decode json " + strValue)
	}

	mapInterface := rawJsonInterface.(map[string]interface{})
	frameDatums := make(map[string]FrameDatum)

	// TODO: is there a better way to do this?
	for k, v := range mapInterface {
		switch vv := v.(type) {
		case map[string]interface{}:
			fd := FrameDatum{}
			for datumName, datumValue := range vv {
				dName := reflect.ValueOf(datumName).String()
				dValue := reflect.ValueOf(datumValue).String()

				if dName == "s" {
					fd.s = dValue
				} else if dName == "a" {
					fd.a = dValue
				} else if dName == "r" {
					fd.r = dValue
				} else if dName == "ha" {
					fd.ha = dValue
				} else if dName == "ba" {
					fd.ba = dValue
				}
			}
			frameDatums[k] = fd
		default:
			character.Log("Unable to read JSON data")
		}
	}

	return frameDatums
}

func (character *Character) PrintFormattedDatum(name string) string {
	frameDatum := character.frames[name]
	return "startup: " + frameDatum.s + ", active: " + frameDatum.a + ", recovery: " + frameDatum.r + ", adv. on hit: " + frameDatum.ha + ", adv. on block: " + frameDatum.ba
}
