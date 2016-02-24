package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/textproto"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type FrameService struct {
	host  *string
	title *string
}

type Bot struct {
	conn    net.Conn
	host    string
	port    string
	channel string
	nick    string
	pass    string
	logger  *log.Logger
	timeout int
	service *FrameService
}

// TODO: make this use the config file + cmd line args
func NewBot(channel *string, nick *string, pass *string, host *string, title *string) *Bot {
	return &Bot{
		conn:    nil,
		host:    "irc.twitch.tv",
		port:    "6667",
		channel: *channel,
		nick:    *nick,
		pass:    *pass,
		logger:  log.New(os.Stdout, "bot: ", log.Ltime),
		timeout: 60,
		service: &FrameService{host: host, title: title},
	}
}

func (bot *Bot) GetOrigin() string {
	return bot.host + ":" + bot.port
}

func Now() string {
	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)
	return timestamp
}

func (bot *Bot) Log(msg string, vars ...string) {
	timestamp := Now()
	bot.logger.Println(timestamp + " - " + bot.nick + " - " + msg + " " + strings.Join(vars, " "))
}

// http://stackoverflow.com/questions/17640360/file-or-line-similar-in-golang
func fileLine() string {
	_, fileName, fileLine, ok := runtime.Caller(1)
	var s string
	if ok {
		s = fmt.Sprintf("%s:%d", fileName, fileLine)
	} else {
		s = ""
	}
	return s
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
	fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :"+"["+Now()+"] "+"This bot gives frame data for USF4. Type !help for usage"+"\r\n")
}

func (bot *Bot) Message(message string) {
	if message == "" {
		return
	}
	bot.Log("Message ~ " + message)
	fmt.Fprintf(bot.conn, "PRIVMSG "+bot.channel+" :"+message+"\r\n")
}

func main() {
	var (
		channel  = flag.String("channel", "#fgframebot", "Channel bot will join")
		nick     = flag.String("nick", "fgframebot", "Nickname in channel")
		apiUri   = flag.String("api", "http://localhost:8080", "Define service that responds with frame data")
		title    = flag.String("title", "usf4", "Define title to scope frame data")
		passPath = flag.String("botpass", "bot_pass.txt", "Path to Twitch OAuth password file")
	)
	flag.Parse()

	filePass, err := ioutil.ReadFile(*passPath)
	if err != nil {
		log.New(os.Stdout, "error: ", log.Lshortfile).Println("Unable to read bot_pass.txt file")
		os.Exit(1)
	}
	pass := strings.Replace(string(filePass), "\n", "", 0)
	bot := NewBot(channel, nick, &pass, apiUri, title)
	bot.Log("Initialized fgframebot.go")
	bot.Log("Using API endpoint: " + *bot.service.host + "/" + *bot.service.title)
	bot.Connect()
	bot.JoinChannel()

	defer bot.conn.Close()

	reader := bufio.NewReader(bot.conn)
	tp := textproto.NewReader(reader)

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
