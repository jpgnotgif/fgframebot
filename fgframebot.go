package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"time"
)

type Bot struct {
	conn    net.Conn
	host    string
	port    string
	channel string
	nick    string
	pass    string
	logger  *log.Logger
	timeout int
}

// TODO: make this use the config file + cmd line args
func NewBot() *Bot {
	return &Bot{
		conn:    nil,
		host:    "irc.twitch.tv",
		port:    "6667",
		channel: "#fgframebot",
		nick:    "fgframebot",
		pass:    "oauth:cahrwuknwcjxxtxbhtirpfky12iivx",
		logger:  log.New(os.Stdout, "logger: ", log.Lshortfile),
		timeout: 60,
	}
}

func (bot *Bot) GetOrigin() string {
	return bot.host + ":" + bot.port
}

func (bot *Bot) Log(msg string, vars ...string) {
	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)
	bot.logger.Println(timestamp + " - " + bot.nick + " - " + msg + " " + strings.Join(vars, " "))
}

func (bot *Bot) Connect() {
	var err error
	origin := bot.GetOrigin()
	bot.Log("Attempting to connect to", origin)
	bot.conn, err = net.Dial("tcp", bot.GetOrigin())
	if err != nil {
		bot.Log("Failed to connect to IRC server! Reconnecting in " + strconv.Itoa(bot.timeout) + "seconds")
		time.Sleep(time.Duration(bot.timeout) * time.Second)
		bot.Connect()
	}
	bot.Log("Connected ~ [host:port] - " + origin + " : " + "[channel] - " + bot.channel + " : " + "[nick] - " + bot.nick)
}

func (bot *Bot) JoinChannel() {
	bot.Log("Joining ~ NICK: " + bot.nick + ", CHANNEL: " + bot.channel)
	fmt.Fprintf(bot.conn, "USER %s 8 * :%s\r\n", bot.nick, bot.nick)
	fmt.Fprintf(bot.conn, "PASS %s\r\n", bot.pass)
	fmt.Fprintf(bot.conn, "NICK %s\r\n", bot.nick)
	fmt.Fprintf(bot.conn, "JOIN %s\r\n", bot.channel)
	fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :"+"[fgframebot] joined "+bot.channel+" This bot gives frame data for USF4. Type !help for usage"+"\r\n")
}

func (bot *Bot) Message(message string) {
	if message == "" {
		return
	}
	bot.Log("Message ~ " + message)
	fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :"+message+"\r\n")
}

// TODO: write ban & timeout bot commands
func (bot *Bot) ConsoleInput() {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		if text == "/quit" {
			bot.conn.Close()
			os.Exit(0)
		}
		bot.Log("TEXT: " + text)
		bot.Message(text)
	}
}

func main() {
	bot := NewBot()
	//origin := bot.GetOrigin()

	bot.Log("Initialized fgframebot.go")
	bot.Connect()
	bot.JoinChannel()

	defer bot.conn.Close()

	reader := bufio.NewReader(bot.conn)
	tp := textproto.NewReader(reader)

	go bot.ConsoleInput()

	for {
		line, err := tp.ReadLine()
		if err != nil {
			break
		}
		if strings.Contains(line, "PING") {
			pongdata := strings.Split(line, "PING ")
			fmt.Fprintf(bot.conn, "PONG %s\r\n", pongdata[1])
		} else if strings.Contains(line, ".tmi.twitch.tv JOIN "+bot.channel) {
			userjoindata := strings.Split(line, ".tmi.twitch.tv JOIN "+bot.channel)
			userjoined := strings.Split(userjoindata[0], "@")
			bot.Log("USER JOIN ~ " + userjoined[1])
		} else if strings.Contains(line, ".tmi.twitch.tv PRIVMSG "+bot.channel) {
			userdata := strings.Split(line, ".tmi.twitch.tv PRIVMSG "+bot.channel)
			username := strings.Split(userdata[0], "@")
			usermessage := strings.Replace(userdata[1], " :", "", 1)

			bot.Log("USER:MSG ~ " + username[1] + " : " + usermessage)

			bot.ReadCmd(username[1], usermessage)
		}
	}
}
