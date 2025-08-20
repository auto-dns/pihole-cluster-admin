package piholeservice

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
	"github.com/auto-dns/pihole-cluster-admin/internal/transport/httpx"
	"github.com/rs/zerolog"
)

type Service struct {
	cluster     cluster
	piholeStore piholeStore
	logger      zerolog.Logger
}

func NewService(cluster cluster, piholeStore piholeStore, logger zerolog.Logger) *Service {
	return &Service{
		cluster:     cluster,
		piholeStore: piholeStore,
		logger:      logger,
	}
}

func (s *Service) GetAll() ([]*domain.PiholeNode, error) {
	return s.piholeStore.GetAllPiholeNodes()
}

func (s *Service) Add(ctx context.Context, params store.AddPiholeParams) (*domain.PiholeNode, error) {
	insertedNode, err := s.piholeStore.AddPiholeNode(params)
	if err != nil {
		return nil, parseSqlError(err)
	}

	nodeSecret, err := s.piholeStore.GetPiholeNodeSecret(insertedNode.Id)
	if err != nil {
		return nil, err
	}

	cfg := &pihole.ClientConfig{
		Id:       insertedNode.Id,
		Name:     insertedNode.Name,
		Scheme:   insertedNode.Scheme,
		Host:     insertedNode.Host,
		Port:     insertedNode.Port,
		Password: nodeSecret.Password,
	}
	client := pihole.NewClient(cfg, s.logger)
	err = s.cluster.AddClient(ctx, client)
	if err != nil {
		return nil, err
	}

	return insertedNode, nil
}

func (s *Service) Update(ctx context.Context, id int64, params store.UpdatePiholeParams) (*domain.PiholeNode, error) {
	updatedNode, err := s.piholeStore.UpdatePiholeNode(id, params)
	if err != nil {
		return nil, parseSqlError(err)
	}

	nodeSecret, err := s.piholeStore.GetPiholeNodeSecret(updatedNode.Id)
	if err != nil {
		return nil, err
	}

	// Update client in cluster
	cfg := &pihole.ClientConfig{
		Id:       updatedNode.Id,
		Name:     updatedNode.Name,
		Scheme:   updatedNode.Scheme,
		Host:     updatedNode.Host,
		Port:     updatedNode.Port,
		Password: nodeSecret.Password,
	}
	err = s.cluster.UpdateClient(ctx, cfg.Id, cfg)
	if err != nil {
		return nil, err
	}

	return updatedNode, nil
}

func (s *Service) Remove(ctx context.Context, id int64) (bool, error) {
	found, err := s.piholeStore.RemovePiholeNode(id)
	if err != nil {
		return false, err
	}

	if s.cluster.HasClient(ctx, id) {
		err = s.cluster.RemoveClient(ctx, id)
		if err != nil {
			return false, nil
		}
	}

	return found, nil
}

func (s *Service) TestInstanceConnection(ctx context.Context, params TestInstanceConnectionParams) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
		Timeout: 4 * time.Second,
	}

	piholeConfig := &pihole.ClientConfig{
		Id: -1, Name: "",
		Scheme: params.Scheme, Host: params.Host, Port: params.Port, Password: params.Password,
	}
	testClient := pihole.NewClient(piholeConfig, s.logger, pihole.WithHTTPClient(httpClient))

	// Login
	if err := testClient.Login(ctx); err != nil {
		return err
	}

	// Logout
	if err := testClient.Logout(ctx); err != nil {
		s.logger.Warn().Err(err).Msg("error logging out of test pihole client")
	}
	httpClient.CloseIdleConnections()

	return nil
}

func (s *Service) TestExistingConnection(ctx context.Context, id int64, params TestExistingConnectionParams) error {
	// Load client from store
	node, err := s.piholeStore.GetPiholeNode(id)
	if err != nil {
		return err
	}
	nodeSecret, err := s.piholeStore.GetPiholeNodeSecret(id)
	if err != nil {
		return err
	}

	// Merge overrides with existing record
	scheme := node.Scheme
	host := node.Host
	port := node.Port
	pass := nodeSecret.Password

	if params.Scheme != nil {
		scheme = strings.ToLower(strings.TrimSpace(*params.Scheme))
	}
	if params.Host != nil {
		host = strings.TrimSpace(*params.Host)
	}
	if params.Port != nil {
		port = *params.Port
	}
	if params.Password != nil && strings.TrimSpace(*params.Password) != "" {
		pass = *params.Password
	}

	// Create a new temporary test client
	httpClient := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyFromEnvironment, DisableKeepAlives: true},
		Timeout:   4 * time.Second,
	}
	cfg := &pihole.ClientConfig{Id: id, Name: node.Name, Scheme: scheme, Host: host, Port: port, Password: pass}
	testClient := pihole.NewClient(cfg, s.logger, pihole.WithHTTPClient(httpClient))

	// Log in
	if err := testClient.Login(ctx); err != nil {
		return err
	}
	// Log out
	if err := testClient.Logout(ctx); err != nil {
		s.logger.Warn().Err(err).Msg("test logout error")
	}
	httpClient.CloseIdleConnections()

	return nil
}

func parseSqlError(err error) error {
	if strings.Contains(err.Error(), "piholes.host") {
		return httpx.NewHttpError(httpx.ErrValidation, "duplicate host:port")
	} else if strings.Contains(err.Error(), "piholes.name") {
		return httpx.NewHttpError(httpx.ErrValidation, "duplicate name")
	} else {
		return err
	}
}
