package service

import (
	"bufio"
	"chat/config"
	"chat/pkg"
	"context"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	host   string
	port   string
}

func NewService(logger *zap.Logger, cfg config.Config) *Service {
	logger = logger.With(zap.String("id", uuid.New().String()))

	return &Service{
		logger: logger,
		host:   cfg.Host,
		port:   cfg.Port,
	}
}

func (s *Service) Run(ctx context.Context, msgWritter *bufio.Writer, nick, room string) {
	p2pAddr := fmt.Sprintf("/ip4/%s/tcp/%s", s.host, s.port)
	host, err := libp2p.New(libp2p.ListenAddrStrings(p2pAddr))
	if err != nil {
		s.logger.Fatal("failed to create libp2p Host",
			zap.Error(err),
		)
	}
	s.logger = s.logger.With(zap.Any("p2p-host", host))

	for _, addr := range host.Addrs() {
		fmt.Printf("Server listening at: %s/p2p/%s\n", addr, host.ID())
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		s.logger.Fatal("failed to create PubSub service",
			zap.Any("p2p-host", host),
			zap.Error(err),
		)
	}
	if err := setupDiscovery(host); err != nil {
		s.logger.Fatal("failed to setup mDNS discovery",
			zap.Any("p2p-host", host),
			zap.Error(err),
		)
	}

	errCh := make(chan error)
	defer close(errCh)

	cr := pkg.JoinChatRoom(ctx, ps, host.ID(), nick, room, msgWritter, errCh)
	go func() {
		for err := range errCh {
			s.logger.Error("failed to join in chat room",
				zap.Any("p2p-host", host),
				zap.String("nick", nick),
				zap.String("room", room),
				zap.Error(err),
			)
		}
	}()

	s.sendMessage(cr)
}

func (s *Service) sendMessage(cr *pkg.ChatRoom) {
	logger := s.logger.With(zap.String("nick", cr.Nick), zap.String("room", cr.Room))

	scanner := bufio.NewScanner(os.Stdin)
	for {
		if scanner.Scan() {
			message := scanner.Text()
			if message == "" {
				continue
			}

			logger.Info("received message", zap.String("message", message))

			err := cr.Publish(message)
			if err != nil {
				logger.Error("failed to send message",
					zap.String("message", message),
					zap.Error(err),
				)
				continue
			}

			logger.Info("message was published",
				zap.String("message", message),
			)
		} else {
			if err := scanner.Err(); err != nil {
				logger.Error("failed to get scanner error", zap.Error(err))
			}
			break
		}
	}
}
