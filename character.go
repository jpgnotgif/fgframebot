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

type CharacterError struct {
	errMsg string
	line   string
}

func (characterError *CharacterError) Error() string {
	return characterError.errMsg + "\n" + characterError.line
}

type Character struct {
	name        string
	endpoint    string
	frames      map[string]FrameDatum
	movelistUrl string
	logger      *log.Logger
	err         *CharacterError
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
	SetData(character, bot)
	return character
}

func SetData(character *Character, bot *Bot) {
	uri := *bot.service.host + "/" + *bot.service.title + "/" + character.name

	character.Log("Fetching data for " + character.name)
	frameDatums := make(map[string]FrameDatum)

	resp, err := http.Get(uri)

	if err != nil {
		character.Log("Failed to fetch data for " + character.name)
		character.err = &CharacterError{err.Error(), fileLine()}
		return
	}

	defer resp.Body.Close()
	body, httpReadErr := ioutil.ReadAll(resp.Body)

	if httpReadErr != nil {
		character.Log("Failed to read response body")
		character.err = &CharacterError{err.Error(), fileLine()}
		return
	}

	strValue := string(body[:])

	var rawJsonInterface interface{}
	jsonErr := json.Unmarshal(body, &rawJsonInterface)

	if jsonErr != nil {
		character.Log("Failed to decode json " + strValue)
		character.err = &CharacterError{jsonErr.Error(), fileLine()}
		return
	}

	mapInterface := rawJsonInterface.(map[string]interface{})

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
		case string:
			character.movelistUrl = vv
		default:
			character.Log("Unable to read JSON data")
		}
	}
	character.frames = frameDatums
}

func (character *Character) PrintFormattedDatum(name string) string {
	frameDatum := character.frames[name]
	return "[" + character.name + ":" + name + "] - " + "startup: " + frameDatum.s + ", active: " + frameDatum.a + ", recovery: " + frameDatum.r + ", hit adv: " + frameDatum.ha + ", block adv: " + frameDatum.ba
}

func (character *Character) PrintFormattedMoveList() string {
	return "[" + character.name + "] - " + character.movelistUrl
}
