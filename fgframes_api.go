package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

type CharactersResponse struct {
	Names []string
}

type ApiConfig struct {
	endpoint string
}

func (bot *Bot) ReadCmd(userName string, userMessage string) {
	apiConfig := &ApiConfig{endpoint: "http://localhost:8080/usf4"}

	if strings.Contains(userMessage, "!frames") {
		framesCmd := strings.Split(userMessage, "!frames")

		data := strings.Split(framesCmd[1], ":")

		if len(data) != 2 {
			bot.Message("!frames <character>:<move-name>")
			return
		}

		characterName := strings.ToLower(strings.TrimSpace(data[0]))
		characterMove := strings.ToLower(strings.TrimSpace(data[1]))

		character := newCharacter(characterName, apiConfig.endpoint, bot)
		message := character.PrintFormattedDatum(characterMove)

		bot.Message(message)

	} else if strings.Contains(userMessage, "!characters") {
		characters := GetCharacters(bot)
		bot.Message("[character list] " + characters)

	} else if strings.Contains(userMessage, "!moves") {
		movesCmd := strings.Split(userMessage, "!moves")

		bot.Log(strconv.Itoa(len(movesCmd)))

		characterName := strings.ToLower(strings.TrimSpace(movesCmd[1]))
		character := newCharacter(characterName, apiConfig.endpoint, bot)
		message := character.PrintFormattedMoveList()

		bot.Message(message)

	} else {
		bot.Message("Command not supported BibleThump")
	}
}

func GetCharacters(bot *Bot) string {
	uri := *bot.service.host + "/" + *bot.service.title + "/" + "characters"

	bot.Log("Fetching characters from " + uri)

	resp, err := http.Get(uri)

	if err != nil {
		// handle error here
		bot.Log("Failed to request list of characters")
		return ""
	}

	defer resp.Body.Close()
	body, httpReadErr := ioutil.ReadAll(resp.Body)

	if httpReadErr != nil {
		bot.Log("Failed to read response body from " + uri)
		return ""
	}

	charactersResp := CharactersResponse{}
	jsonErr := json.Unmarshal(body, &charactersResp)

	if jsonErr != nil {
		bot.Log("Failed to parse JSON response from " + uri)
		return ""
	}
	names := charactersResp.Names
	return strings.Join(names, ", ")
}
