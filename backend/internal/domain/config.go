package domain

// PiholeConfig holds the subset of Pi-hole config fields managed by the cluster admin.
// All fields use pointers so callers can distinguish "not set" from zero value when
// building partial PATCH requests.
type PiholeConfig struct {
	DNS       PiholeConfigDNS
	Misc      PiholeConfigMisc
	FTL       PiholeConfigFTL
	Webserver PiholeConfigWebserver
	Resolver  PiholeConfigResolver
}

type PiholeConfigDNS struct {
	Upstreams    []string
	Interface    string
	Port         int
	DNSSEC       bool
	DomainNeeded bool
	ExpandHosts  bool
	LocalTTL     int
	BlockingMode string
	BlockingIPv4 string
	BlockingIPv6 string
	RateLimit    PiholeConfigRateLimit
	RevServer    PiholeConfigRevServer
	PiholePTR    string
	QueryLog     PiholeConfigQueryLog
}

type PiholeConfigRateLimit struct {
	Count    int
	Interval int
}

type PiholeConfigRevServer struct {
	Active bool
	CIDR   string
	Target string
	Domain string
}

type PiholeConfigQueryLog struct {
	Enabled bool
}

type PiholeConfigMisc struct {
	PrivacyLevel int
	DelayStartup int
	Nice         int
}

type PiholeConfigFTL struct {
	QueryDisplay string
	DelayStartup int
	Database     PiholeConfigFTLDatabase
}

type PiholeConfigFTLDatabase struct {
	DBInterval float64
	MaxDBDays  int
}

type PiholeConfigWebserver struct {
	API PiholeConfigWebserverAPI
}

type PiholeConfigWebserverAPI struct {
	ExcludeClients []string
	ExcludeDomains []string
	MaxHistory     int
}

type PiholeConfigResolver struct {
	ResolveIPv4  bool
	ResolveIPv6  bool
	NetworkNames bool
}

// NodeConfigResult is a per-node config read result.
type NodeConfigResult struct {
	Config *PiholeConfig
}

// ClusterConfig aggregates per-node configs with drift detection.
type ClusterConfig struct {
	// Consensus holds the value from the first node (or majority); used as the
	// displayed value when no drift. Nil if all nodes errored.
	Consensus *PiholeConfig
	// PerNode holds each node's raw config for drift computation on the frontend.
	PerNode map[int64]*PiholeConfig
	// Drifted is true if any field differs across nodes.
	Drifted bool
}

// PiholeConfigPatch is the set of fields the caller wants to change.
// Each field is a pointer; nil means "leave unchanged."
type PiholeConfigPatch struct {
	DNS       *PiholeConfigDNSPatch
	Misc      *PiholeConfigMiscPatch
	FTL       *PiholeConfigFTLPatch
	Webserver *PiholeConfigWebserverPatch
	Resolver  *PiholeConfigResolverPatch
}

type PiholeConfigDNSPatch struct {
	Upstreams    *[]string
	Interface    *string
	Port         *int
	DNSSEC       *bool
	DomainNeeded *bool
	ExpandHosts  *bool
	LocalTTL     *int
	BlockingMode *string
	BlockingIPv4 *string
	BlockingIPv6 *string
	RateLimit    *PiholeConfigRateLimit
	RevServer    *PiholeConfigRevServer
	PiholePTR    *string
	QueryLog     *PiholeConfigQueryLogPatch
}

type PiholeConfigQueryLogPatch struct {
	Enabled *bool
}

type PiholeConfigMiscPatch struct {
	PrivacyLevel *int
	DelayStartup *int
	Nice         *int
}

type PiholeConfigFTLPatch struct {
	QueryDisplay *string
	DelayStartup *int
	Database     *PiholeConfigFTLDatabasePatch
}

type PiholeConfigFTLDatabasePatch struct {
	DBInterval *float64
	MaxDBDays  *int
}

type PiholeConfigWebserverPatch struct {
	API *PiholeConfigWebserverAPIPatch
}

type PiholeConfigWebserverAPIPatch struct {
	ExcludeClients *[]string
	ExcludeDomains *[]string
	MaxHistory     *int
}

type PiholeConfigResolverPatch struct {
	ResolveIPv4  *bool
	ResolveIPv6  *bool
	NetworkNames *bool
}
