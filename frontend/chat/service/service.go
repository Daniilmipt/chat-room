package service

import (
	"chatroom/chat/config"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/net/connmgr"
	"github.com/pkg/errors"
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

func (s *Service) GetPubSub(ctx context.Context) (*pubsub.PubSub, host.Host, error) {
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519,
		-1,
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to generate key pair")
	}

	connmgr, err := connmgr.NewConnManager(
		100,
		400,
		connmgr.WithGracePeriod(time.Hour*24),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create connection manager")
	}

	p2pAddr := fmt.Sprintf("/ip4/%s/tcp/%s", s.host, s.port)

	host, err := libp2p.New(
		libp2p.Identity(priv),
		libp2p.ListenAddrStrings(p2pAddr),
		libp2p.ConnectionManager(connmgr),
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err := dht.New(ctx, h)
			return idht, err
		}),
		libp2p.NATPortMap(),
		libp2p.EnableNATService(),
	)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to create libp2p Host")
	}
	s.logger = s.logger.With(zap.Any("p2p-host", host))

	for _, addr := range dht.DefaultBootstrapPeers {
		pi, err := peer.AddrInfoFromP2pAddr(addr)
		if err != nil {
			s.logger.Error("failed to connect to bootstrap peer", zap.Error(err), zap.String("address", addr.String()))
			continue
		}

		if err := host.Connect(ctx, *pi); err != nil {
			s.logger.Error("failed to connect to bootstrap peer", zap.Error(err), zap.Any("peer", pi))
		}
	}

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		return nil, host, errors.Wrap(err, "failed to create PubSub service")
	}
	if err := setupDiscovery(host); err != nil {
		return ps, host, errors.Wrap(err, "failed to setup mDNS discovery")
	}

	return ps, host, nil
}
