package main

import (
	"bufio"
	"chat/pkg"
	"chat/service"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
)

func parseFlags() (string, string) {
	nickFlag := flag.String("nick", "anonymous", "room to use in chat. will be \"anonymous\" if empty")
	roomFlag := flag.String("room", "main", "name of chat room to join. will be \"main\" if empty")
	flag.Parse()

	room := *roomFlag
	nick := *nickFlag

	return room, nick
}

func messageLogWritter(room string) (*os.File, *bufio.Writer) {
	filepath := fmt.Sprintf("../messages/%s.log", room)
	logFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed to open room log file: %v", err)
	}
	writer := bufio.NewWriter(logFile)

	return logFile, writer
}

func main() {
	room, nick := parseFlags()

	logger, f := pkg.SetupLogger()
	defer f.Close()

	msgF, msgWritter := messageLogWritter(room)
	defer msgF.Close()

	cfg := parseConfig()
	s := service.NewService(logger, cfg.Host, cfg.Port)

	ctx := context.Background()
	s.Run(ctx, msgWritter, nick, room)
}
