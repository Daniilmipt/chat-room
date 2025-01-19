package main

import (
	"bufio"
	"chat/config"
	"chat/pkg"
	"chat/service"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "0"
)

func parseFlags() (string, string, config.Config) {
	nickFlag := flag.String("nick", "anonymous", "room to use in chat. will be \"anonymous\" if empty")
	roomFlag := flag.String("room", "main", "name of chat room to join. will be \"main\" if empty")
	hostFlag := flag.String("host", defaultHost, "host which we will listen p2p")
	portFlag := flag.String("port", defaultPort, "port for listening p2p")
	flag.Parse()

	room := *roomFlag
	nick := *nickFlag
	host := *hostFlag
	port := *portFlag

	return room, nick, config.Config{Host: host, Port: port}
}

func messageLogWritter(room string) (*os.File, *bufio.Writer) {
	filepath := fmt.Sprintf("./messages/%s.log", room)
	logFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open room log file: %v", err)
	}
	writer := bufio.NewWriter(logFile)

	return logFile, writer
}

func main() {
	room, nick, cfg := parseFlags()

	logger, f := pkg.SetupLogger()
	defer f.Close()

	msgF, msgWritter := messageLogWritter(room)
	defer msgF.Close()

	s := service.NewService(logger, cfg)

	ctx := context.Background()
	s.Run(ctx, msgWritter, nick, room)
}
