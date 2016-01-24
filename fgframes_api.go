package main

import (
	"fmt"
	"strings"
)

func (bot *Bot) ReadCmd(userName string, userMessage string) {

	if strings.Contains(userMessage, "!frames") {
		framesCmd := strings.Split(userMessage, "!frames")

		data := strings.Split(framesCmd[1], ":")

		if len(data) != 2 {
			bot.Message("!frames <character>:<move-name>")
			return
		}

		character := strings.Trim(data[0])
		move := strings.Trim(data[1])

		// TODO: Continue here

	} else {
		bot.Message("Command not supported BibleThump")
	}
}

func (bot *Bot) getFrames(character string, move string) {

}
