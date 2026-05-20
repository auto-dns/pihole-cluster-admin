package v1

import "github.com/auto-dns/pihole-cluster-admin/internal/domain"

// GET /config response

type getConfigResponseDTO struct {
	Consensus *configDTO            `json:"consensus"`
	PerNode   map[int64]*configDTO  `json:"perNode"`
	Drifted   bool                  `json:"drifted"`
}

type configDTO struct {
	DNS       configDNSDTO       `json:"dns"`
	Misc      configMiscDTO      `json:"misc"`
	FTL       configFTLDTO       `json:"ftl"`
	Webserver configWebserverDTO `json:"webserver"`
	Resolver  configResolverDTO  `json:"resolver"`
}

type configDNSDTO struct {
	Upstreams    []string              `json:"upstreams"`
	Interface    string                `json:"interface"`
	Port         int                   `json:"port"`
	DNSSEC       bool                  `json:"dnssec"`
	DomainNeeded bool                  `json:"domainNeeded"`
	ExpandHosts  bool                  `json:"expandHosts"`
	LocalTTL     int                   `json:"localTTL"`
	BlockingMode string                `json:"blockingmode"`
	BlockingIPv4 string                `json:"blockingipv4"`
	BlockingIPv6 string                `json:"blockingipv6"`
	RateLimit    configRateLimitDTO    `json:"ratelimit"`
	RevServer    configRevServerDTO    `json:"revServer"`
	PiholePTR    string                `json:"piholePTR"`
	QueryLog     configQueryLogDTO     `json:"querylog"`
}

type configRateLimitDTO struct {
	Count    int `json:"count"`
	Interval int `json:"interval"`
}

type configRevServerDTO struct {
	Active bool   `json:"active"`
	CIDR   string `json:"cidr"`
	Target string `json:"target"`
	Domain string `json:"domain"`
}

type configQueryLogDTO struct {
	Enabled bool `json:"enabled"`
}

type configMiscDTO struct {
	PrivacyLevel int `json:"privacylevel"`
	DelayStartup int `json:"delay_startup"`
	Nice         int `json:"nice"`
}

type configFTLDTO struct {
	QueryDisplay string          `json:"query_display"`
	DelayStartup int             `json:"delay_startup"`
	Database     configFTLDBDTO  `json:"database"`
}

type configFTLDBDTO struct {
	DBInterval float64 `json:"DBinterval"`
	MaxDBDays  int     `json:"maxDBdays"`
}

type configWebserverDTO struct {
	API configWebserverAPIDTO `json:"api"`
}

type configWebserverAPIDTO struct {
	ExcludeClients []string `json:"excludeClients"`
	ExcludeDomains []string `json:"excludeDomains"`
	MaxHistory     int      `json:"maxHistory"`
}

type configResolverDTO struct {
	ResolveIPv4  bool `json:"resolveIPv4"`
	ResolveIPv6  bool `json:"resolveIPv6"`
	NetworkNames bool `json:"networkNames"`
}

// PATCH /config request

type patchConfigRequestDTO struct {
	DNS       *patchConfigDNSDTO       `json:"dns,omitempty"`
	Misc      *patchConfigMiscDTO      `json:"misc,omitempty"`
	FTL       *patchConfigFTLDTO       `json:"ftl,omitempty"`
	Webserver *patchConfigWebserverDTO `json:"webserver,omitempty"`
	Resolver  *patchConfigResolverDTO  `json:"resolver,omitempty"`
}

type patchConfigDNSDTO struct {
	Upstreams    *[]string                  `json:"upstreams,omitempty"`
	Interface    *string                    `json:"interface,omitempty"`
	Port         *int                       `json:"port,omitempty"`
	DNSSEC       *bool                      `json:"dnssec,omitempty"`
	DomainNeeded *bool                      `json:"domainNeeded,omitempty"`
	ExpandHosts  *bool                      `json:"expandHosts,omitempty"`
	LocalTTL     *int                       `json:"localTTL,omitempty"`
	BlockingMode *string                    `json:"blockingmode,omitempty"`
	BlockingIPv4 *string                    `json:"blockingipv4,omitempty"`
	BlockingIPv6 *string                    `json:"blockingipv6,omitempty"`
	RateLimit    *patchConfigRateLimitDTO   `json:"ratelimit,omitempty"`
	RevServer    *patchConfigRevServerDTO   `json:"revServer,omitempty"`
	PiholePTR    *string                    `json:"piholePTR,omitempty"`
	QueryLog     *patchConfigQueryLogDTO    `json:"querylog,omitempty"`
}

type patchConfigRateLimitDTO struct {
	Count    *int `json:"count,omitempty"`
	Interval *int `json:"interval,omitempty"`
}

type patchConfigRevServerDTO struct {
	Active *bool   `json:"active,omitempty"`
	CIDR   *string `json:"cidr,omitempty"`
	Target *string `json:"target,omitempty"`
	Domain *string `json:"domain,omitempty"`
}

type patchConfigQueryLogDTO struct {
	Enabled *bool `json:"enabled,omitempty"`
}

type patchConfigMiscDTO struct {
	PrivacyLevel *int `json:"privacylevel,omitempty"`
	DelayStartup *int `json:"delay_startup,omitempty"`
	Nice         *int `json:"nice,omitempty"`
}

type patchConfigFTLDTO struct {
	QueryDisplay *string              `json:"query_display,omitempty"`
	DelayStartup *int                 `json:"delay_startup,omitempty"`
	Database     *patchConfigFTLDBDTO `json:"database,omitempty"`
}

type patchConfigFTLDBDTO struct {
	DBInterval *float64 `json:"DBinterval,omitempty"`
	MaxDBDays  *int     `json:"maxDBdays,omitempty"`
}

type patchConfigWebserverDTO struct {
	API *patchConfigWebserverAPIDTO `json:"api,omitempty"`
}

type patchConfigWebserverAPIDTO struct {
	ExcludeClients *[]string `json:"excludeClients,omitempty"`
	ExcludeDomains *[]string `json:"excludeDomains,omitempty"`
	MaxHistory     *int      `json:"maxHistory,omitempty"`
}

type patchConfigResolverDTO struct {
	ResolveIPv4  *bool `json:"resolveIPv4,omitempty"`
	ResolveIPv6  *bool `json:"resolveIPv6,omitempty"`
	NetworkNames *bool `json:"networkNames,omitempty"`
}

// PATCH /config response

type patchConfigResponseDTO struct {
	Nodes map[int64]patchConfigNodeDTO `json:"nodes"`
}

type patchConfigNodeDTO struct {
	Node    piholeNodeRefDTO `json:"node"`
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
}

// Converters

func toConfigDTO(c domain.PiholeConfig) configDTO {
	ups := c.DNS.Upstreams
	if ups == nil {
		ups = []string{}
	}
	excClients := c.Webserver.API.ExcludeClients
	if excClients == nil {
		excClients = []string{}
	}
	excDomains := c.Webserver.API.ExcludeDomains
	if excDomains == nil {
		excDomains = []string{}
	}
	return configDTO{
		DNS: configDNSDTO{
			Upstreams:    ups,
			Interface:    c.DNS.Interface,
			Port:         c.DNS.Port,
			DNSSEC:       c.DNS.DNSSEC,
			DomainNeeded: c.DNS.DomainNeeded,
			ExpandHosts:  c.DNS.ExpandHosts,
			LocalTTL:     c.DNS.LocalTTL,
			BlockingMode: c.DNS.BlockingMode,
			BlockingIPv4: c.DNS.BlockingIPv4,
			BlockingIPv6: c.DNS.BlockingIPv6,
			RateLimit: configRateLimitDTO{
				Count:    c.DNS.RateLimit.Count,
				Interval: c.DNS.RateLimit.Interval,
			},
			RevServer: configRevServerDTO{
				Active: c.DNS.RevServer.Active,
				CIDR:   c.DNS.RevServer.CIDR,
				Target: c.DNS.RevServer.Target,
				Domain: c.DNS.RevServer.Domain,
			},
			PiholePTR: c.DNS.PiholePTR,
			QueryLog:  configQueryLogDTO{Enabled: c.DNS.QueryLog.Enabled},
		},
		Misc: configMiscDTO{
			PrivacyLevel: c.Misc.PrivacyLevel,
			DelayStartup: c.Misc.DelayStartup,
			Nice:         c.Misc.Nice,
		},
		FTL: configFTLDTO{
			QueryDisplay: c.FTL.QueryDisplay,
			DelayStartup: c.FTL.DelayStartup,
			Database: configFTLDBDTO{
				DBInterval: c.FTL.Database.DBInterval,
				MaxDBDays:  c.FTL.Database.MaxDBDays,
			},
		},
		Webserver: configWebserverDTO{
			API: configWebserverAPIDTO{
				ExcludeClients: excClients,
				ExcludeDomains: excDomains,
				MaxHistory:     c.Webserver.API.MaxHistory,
			},
		},
		Resolver: configResolverDTO{
			ResolveIPv4:  c.Resolver.ResolveIPv4,
			ResolveIPv6:  c.Resolver.ResolveIPv6,
			NetworkNames: c.Resolver.NetworkNames,
		},
	}
}

func patchRequestToDomain(req patchConfigRequestDTO) domain.PiholeConfigPatch {
	patch := domain.PiholeConfigPatch{}
	if req.DNS != nil {
		d := req.DNS
		dp := &domain.PiholeConfigDNSPatch{
			Upstreams:    d.Upstreams,
			Interface:    d.Interface,
			Port:         d.Port,
			DNSSEC:       d.DNSSEC,
			DomainNeeded: d.DomainNeeded,
			ExpandHosts:  d.ExpandHosts,
			LocalTTL:     d.LocalTTL,
			BlockingMode: d.BlockingMode,
			BlockingIPv4: d.BlockingIPv4,
			BlockingIPv6: d.BlockingIPv6,
			PiholePTR:    d.PiholePTR,
		}
		if d.RateLimit != nil {
			rl := domain.PiholeConfigRateLimit{}
			if d.RateLimit.Count != nil {
				rl.Count = *d.RateLimit.Count
			}
			if d.RateLimit.Interval != nil {
				rl.Interval = *d.RateLimit.Interval
			}
			dp.RateLimit = &rl
		}
		if d.RevServer != nil {
			rs := domain.PiholeConfigRevServer{}
			if d.RevServer.Active != nil {
				rs.Active = *d.RevServer.Active
			}
			if d.RevServer.CIDR != nil {
				rs.CIDR = *d.RevServer.CIDR
			}
			if d.RevServer.Target != nil {
				rs.Target = *d.RevServer.Target
			}
			if d.RevServer.Domain != nil {
				rs.Domain = *d.RevServer.Domain
			}
			dp.RevServer = &rs
		}
		if d.QueryLog != nil {
			dp.QueryLog = &domain.PiholeConfigQueryLogPatch{Enabled: d.QueryLog.Enabled}
		}
		patch.DNS = dp
	}
	if req.Misc != nil {
		patch.Misc = &domain.PiholeConfigMiscPatch{
			PrivacyLevel: req.Misc.PrivacyLevel,
			DelayStartup: req.Misc.DelayStartup,
			Nice:         req.Misc.Nice,
		}
	}
	if req.FTL != nil {
		fp := &domain.PiholeConfigFTLPatch{
			QueryDisplay: req.FTL.QueryDisplay,
			DelayStartup: req.FTL.DelayStartup,
		}
		if req.FTL.Database != nil {
			fp.Database = &domain.PiholeConfigFTLDatabasePatch{
				DBInterval: req.FTL.Database.DBInterval,
				MaxDBDays:  req.FTL.Database.MaxDBDays,
			}
		}
		patch.FTL = fp
	}
	if req.Webserver != nil && req.Webserver.API != nil {
		patch.Webserver = &domain.PiholeConfigWebserverPatch{
			API: &domain.PiholeConfigWebserverAPIPatch{
				ExcludeClients: req.Webserver.API.ExcludeClients,
				ExcludeDomains: req.Webserver.API.ExcludeDomains,
				MaxHistory:     req.Webserver.API.MaxHistory,
			},
		}
	}
	if req.Resolver != nil {
		patch.Resolver = &domain.PiholeConfigResolverPatch{
			ResolveIPv4:  req.Resolver.ResolveIPv4,
			ResolveIPv6:  req.Resolver.ResolveIPv6,
			NetworkNames: req.Resolver.NetworkNames,
		}
	}
	return patch
}
