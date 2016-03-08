# fgframebot
A golang based, Twitch IRC bot that provides fighting game character frame data. This bot relies on the [fgframes API](https://github.com/jpgnotgif/fgframes). This project was built as a means of exploring golang and building an http client that uses an external API.

## Bootstrap
- [Install golang](https://golang.org/doc/install)
- Build and install server
```
$ go install github.com/jpgnotgif/fgframebot
```
- [Create Twitch OAuth token](http://twitchapps.com/tmi/)
- Create file named ***bot_pass.txt*** in $GOPATH/bin
- Run server
```
$ $GOPATH/bin/fgframebot
```
