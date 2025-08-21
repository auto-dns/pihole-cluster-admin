package app

import (
	"github.com/auto-dns/pihole-cluster-admin/internal/pihole"
	"github.com/rs/zerolog"
)

func GetClients(piholeGetter PiholeGetter, logger zerolog.Logger) (map[int64]*pihole.Client, error) {
	// Load piholes from database
	nodes, err := piholeGetter.GetAllPiholeNodes()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load pihole nodes from database")
		return nil, err
	}

	clients := make(map[int64]*pihole.Client, len(nodes))
	for _, node := range nodes {
		node := node

		nodeSecret, err := piholeGetter.GetPiholeNodeSecret(node.Id)
		if err != nil {
			return nil, err
		}

		cfg := &pihole.ClientConfig{
			Id:       node.Id,
			Scheme:   node.Scheme,
			Host:     node.Host,
			Port:     node.Port,
			Password: nodeSecret.Password,
			Name:     node.Name,
		}
		nodeLogger := logger.With().Int64("db_id", node.Id).Str("host", node.Host).Int("port", node.Port).Logger()
		clients[node.Id] = pihole.NewClient(cfg, nodeLogger)
	}
	logger.Info().Int("node_count", len(nodes)).Msg("loaded pihole nodes")

	return clients, nil
}
