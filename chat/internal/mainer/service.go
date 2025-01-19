package mainer

import (
	"bufio"
	"chat/internal/pkg"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/zap"
)

const (
	defaultHost = "0.0.0.0"
	defaultPort = "9090"
)

type Service struct {
	logger *zap.Logger
	host   string
	port   string
}

func NewService(logger *zap.Logger, host, port string) *Service {
	logger = logger.With(zap.String("id", uuid.New().String()))

	if host == "" {
		host = defaultHost
	}
	if port == "" {
		port = defaultPort
	}

	return &Service{
		logger: logger,
		host:   host,
		port:   port,
	}
}

func (s *Service) Run(ctx context.Context, msgWritter *bufio.Writer, nick, room string) {
	p2pAddr := fmt.Sprintf("/ip4/%s/tcp/%s", s.host, s.port)

	s.logger = s.logger.With(zap.String("address", p2pAddr))

	host, err := libp2p.New(libp2p.ListenAddrStrings(p2pAddr))
	if err != nil {
		s.logger.Fatal("failed to create libp2p Host", zap.Error(err))
	}

	for _, addr := range host.Addrs() {
		s.logger.Info("Server listening",
			zap.String("p2p_address", fmt.Sprintf("%s/p2p/%s\n", addr, host.ID())),
		)
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		s.logger.Fatal("failed to create PubSub service", zap.Error(err))
	}
	if err := pkg.SetupDiscovery(host); err != nil {
		s.logger.Fatal("failed to setup mDNS discovery", zap.Error(err))
	}

	errCh := make(chan error)
	defer close(errCh)

	cr := pkg.JoinChatRoom(ctx, ps, host.ID(), nick, room, msgWritter, errCh)
	go func() {
		for err := range errCh {
			s.logger.Error("failed to join in chat room",
				zap.String("nick", nick),
				zap.String("room", room),
				zap.Error(err),
			)
		}
	}()

	pkg.SendMessage(cr, s.logger)
}
