package main

import (
	"strings"
)

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

		bot.Log(character.name)

		message := character.PrintFormattedDatum(characterMove)

		bot.Message(message)

	} else {
		bot.Message("Command not supported BibleThump")
	}
}
