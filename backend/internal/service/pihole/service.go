package piholesvc

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/auto-dns/pihole-cluster-admin/internal/store"
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

func (s *Service) Add(ctx context.Context, cmd AddNodeCommand) (*domain.PiholeNode, error) {
	params := store.AddPiholeParams{
		Scheme:      cmd.Scheme,
		Host:        cmd.Host,
		Port:        cmd.Port,
		Name:        cmd.Name,
		Description: cmd.Description,
		Password:    cmd.Password,
	}

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

func (s *Service) Update(ctx context.Context, id int64, cmd UpdateNodeCommand) (*domain.PiholeNode, error) {
	params := store.UpdatePiholeParams{
		Scheme:      cmd.Scheme,
		Host:        cmd.Host,
		Port:        cmd.Port,
		Name:        cmd.Name,
		Description: cmd.Description,
		Password:    cmd.Password,
	}

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

	if s.cluster.HasClient(ctx, updatedNode.Id) {
		err = s.cluster.UpdateClient(ctx, updatedNode.Id, cfg)
	} else {
		err = s.cluster.AddClient(ctx, pihole.NewClient(cfg, s.logger))
	}
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

func (s *Service) TestInstanceConnection(ctx context.Context, cmd TestInstanceConnectionCommand) error {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
		Timeout: 4 * time.Second,
	}

	piholeConfig := &pihole.ClientConfig{
		Id:       -1,
		Name:     "",
		Scheme:   cmd.Scheme,
		Host:     cmd.Host,
		Port:     cmd.Port,
		Password: cmd.Password,
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

func (s *Service) TestExistingConnection(ctx context.Context, id int64, cmd TestExistingConnectionCommand) error {
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

	if cmd.Scheme != nil {
		scheme = strings.ToLower(strings.TrimSpace(*cmd.Scheme))
	}
	if cmd.Host != nil {
		host = strings.TrimSpace(*cmd.Host)
	}
	if cmd.Port != nil {
		port = *cmd.Port
	}
	if cmd.Password != nil && strings.TrimSpace(*cmd.Password) != "" {
		pass = *cmd.Password
	}

	// Create a new temporary test client
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:             http.ProxyFromEnvironment,
			DisableKeepAlives: true,
		},
		Timeout: 4 * time.Second,
	}
	cfg := &pihole.ClientConfig{
		Id:       id,
		Name:     node.Name,
		Scheme:   scheme,
		Host:     host,
		Port:     port,
		Password: pass,
	}
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

func (s *Service) RotatePassword(ctx context.Context, id int64, cmd RotatePasswordCommand) error {
	// Push new password to the Pi-hole node via PATCH /api/config.
	// This must happen first — if it fails, the stored credential is not touched.
	if err := s.cluster.SetPasswordForNode(ctx, id, cmd.NewPassword); err != nil {
		return err
	}

	// Update stored credential in DB and refresh the cluster client.
	// If this fails, Pi-hole already has the new password but cluster admin still
	// holds the old credential. The operator must update the stored credential
	// manually via the node edit form.
	pass := cmd.NewPassword
	updatedNode, err := s.piholeStore.UpdatePiholeNode(id, store.UpdatePiholeParams{Password: &pass})
	if err != nil {
		s.logger.Error().Err(err).Int64("nodeId", id).Msg("Pi-hole password changed but stored credential update failed — update the credential manually via the node edit form")
		return fmt.Errorf("Pi-hole password changed successfully but the stored credential could not be saved (%w) — update it manually via the node edit form", err)
	}

	nodeSecret, err := s.piholeStore.GetPiholeNodeSecret(updatedNode.Id)
	if err != nil {
		return err
	}

	cfg := &pihole.ClientConfig{
		Id:       updatedNode.Id,
		Name:     updatedNode.Name,
		Scheme:   updatedNode.Scheme,
		Host:     updatedNode.Host,
		Port:     updatedNode.Port,
		Password: nodeSecret.Password,
	}

	if s.cluster.HasClient(ctx, updatedNode.Id) {
		return s.cluster.UpdateClient(ctx, updatedNode.Id, cfg)
	}
	return s.cluster.AddClient(ctx, pihole.NewClient(cfg, s.logger))
}

func parseSqlError(err error) error {
	if strings.Contains(err.Error(), "piholes.host") {
		return errs.Invalid("duplicate host:port", err)
	} else if strings.Contains(err.Error(), "piholes.name") {
		return errs.Invalid("duplicate name", err)
	} else {
		return err
	}
}
