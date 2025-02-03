package main

import (
	"bufio"
	"chat/config"
	"chat/internal/node"
	"chat/internal/peer"
	"chat/internal/pkg"
	"context"
	"flag"
	"fmt"
	"os"

	"go.uber.org/zap"
)

func parseFlags() (string, string, bool, config.Config) {
	var (
		nick, room, host, port, peerID string
		isNode                         bool
	)

	flag.StringVar(&nick, "nick", "anonymous", "room to use in chat. must be not empty")
	flag.StringVar(&room, "room", "main", "name of chat room to join. must be not empty")
	flag.StringVar(&host, "host", "", "host which we will listen p2p")
	flag.StringVar(&port, "port", "", "port for listening p2p")
	flag.StringVar(&peerID, "peerid", "", "peerID for listening p2p")
	flag.BoolVar(&isNode, "node", false, "do we run chat as a node, which you can connect to")

	flag.Parse()

	return room, nick, isNode, config.Config{Host: host, Port: port, PeerID: peerID}
}

func messageLogWritter(room string, logger *zap.Logger) (*os.File, *bufio.Writer) {
	if err := os.Mkdir("./messages", os.ModePerm); err != nil {
		logger.Info("can not create messages directory", zap.Error(err))
	}

	filepath := fmt.Sprintf("./messages/%s.log", room)
	logFile, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		logger.Error("failed to open room log file", zap.Error(err))
		return nil, nil
	}
	writer := bufio.NewWriter(logFile)

	return logFile, writer
}

type ChatService interface {
	Run(ctx context.Context, msgWritter *bufio.Writer, nick, room string)
}

func main() {
	room, nick, isNode, cfg := parseFlags()

	logger, f := pkg.SetupLogger()
	defer f.Close()

	msgF, msgWritter := messageLogWritter(room, logger)
	defer msgF.Close()

	var s ChatService
	if isNode {
		s = node.NewService(logger, cfg.Host, cfg.Port)
	} else {
		s = peer.NewService(logger, cfg.Host, cfg.Port, cfg.PeerID)
	}

	ctx := context.Background()
	s.Run(ctx, msgWritter, nick, room)
}
