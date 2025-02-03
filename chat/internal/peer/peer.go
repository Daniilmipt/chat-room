package peer

import (
	"bufio"
	"chat/internal/pkg"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.Logger
	host   string
	port   string
	peerID string
}

func NewService(logger *zap.Logger, host, port, peerID string) *Service {
	logger = logger.With(zap.String("id", uuid.New().String()))

	if host == "" || port == "" || peerID == "" {
		logger.Error("invalid host, port or peerID",
			zap.String("address", fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", host, port, peerID)),
		)
	}

	return &Service{
		logger: logger,
		host:   host,
		port:   port,
		peerID: peerID,
	}
}

func (s *Service) initConnection(ctx context.Context, p2pAddress string) (*host.Host, error) {
	host, err := libp2p.New()
	if err != nil {
		return nil, fmt.Errorf("failed to create libp2p Host: %s", err)
	}

	serverMA, err := multiaddr.NewMultiaddr(p2pAddress)
	if err != nil {
		return nil, fmt.Errorf("invalid server address: %s", err)
	}

	serverInfo, err := peer.AddrInfoFromP2pAddr(serverMA)
	if err != nil {
		return nil, fmt.Errorf("failed to parse server address: %s", err)
	}

	if err := host.Connect(ctx, *serverInfo); err != nil {
		return nil, fmt.Errorf("failed to connect to server: %s", err)
	}

	return &host, nil
}

func (s *Service) Run(ctx context.Context, msgWritter *bufio.Writer, nick, room string) {
	p2pAddr := fmt.Sprintf("/ip4/%s/tcp/%s/p2p/%s", s.host, s.port, s.peerID)

	s.logger = s.logger.With(zap.String("address", p2pAddr))

	host, err := s.initConnection(ctx, p2pAddr)
	if err != nil {
		s.logger.Fatal("failed to get connection to peer", zap.Error(err))
	}

	s.logger.Info("connected to server")

	ps, err := pubsub.NewGossipSub(ctx, *host)
	if err != nil {
		s.logger.Fatal("failed to create PubSub service", zap.Error(err))
	}
	if err := pkg.SetupDiscovery(*host); err != nil {
		s.logger.Fatal("failed to setup mDNS discovery", zap.Error(err))
	}

	errCh := make(chan error)
	defer close(errCh)

	cr := pkg.JoinChatRoom(ctx, ps, (*host).ID(), nick, room, msgWritter, errCh)
	go func() {
		for err := range errCh {
			s.logger.Error("failed to join in chat room",
				zap.Any("p2p-host", *host),
				zap.String("nick", nick),
				zap.String("room", room),
				zap.Error(err),
			)
		}
	}()

	pkg.SendMessage(cr, s.logger)
}
