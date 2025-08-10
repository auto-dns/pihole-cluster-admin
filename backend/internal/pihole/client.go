package pihole

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

func buildQueryParams(req FetchQueryLogClientRequest) string {
	params := url.Values{}

	// Pagination
	if req.Cursor != nil {
		params.Set("cursor", fmt.Sprintf("%d", *req.Cursor))
	}
	if req.Length != nil {
		params.Set("length", fmt.Sprintf("%d", *req.Length))
	}
	if req.Start != nil {
		params.Set("start", fmt.Sprintf("%d", *req.Start))
	}

	// Filters
	f := req.Filters
	if f.From != nil {
		params.Set("from", fmt.Sprintf("%d", *f.From))
	}
	if f.Until != nil {
		params.Set("until", fmt.Sprintf("%d", *f.Until))
	}
	if f.Domain != nil {
		params.Set("domain", *f.Domain)
	}
	if f.ClientIP != nil {
		params.Set("client_ip", *f.ClientIP)
	}
	if f.ClientName != nil {
		params.Set("client_name", *f.ClientName)
	}
	if f.Upstream != nil {
		params.Set("upstream", *f.Upstream)
	}
	if f.Type != nil {
		params.Set("type", *f.Type)
	}
	if f.Status != nil {
		params.Set("status", *f.Status)
	}
	if f.Reply != nil {
		params.Set("reply", *f.Reply)
	}
	if f.DNSSEC != nil {
		params.Set("dnssec", *f.DNSSEC)
	}
	if f.Disk != nil {
		params.Set("disk", strconv.FormatBool(*f.Disk))
	}

	return params.Encode()
}

type sessionState struct {
	SID        string
	ValidUntil time.Time
}

type Client struct {
	cfg     *ClientConfig
	HTTP    *http.Client
	session sessionState
	mu      sync.Mutex
	logger  zerolog.Logger
	cfgMu   sync.RWMutex
}

type ClientConfig struct {
	Id       int64
	Name     string
	Scheme   string
	Host     string
	Port     int
	Password string
}

func NewClient(cfg *ClientConfig, logger zerolog.Logger) ClientInterface {
	return &Client{
		cfg:    cfg,
		logger: logger,
		HTTP:   &http.Client{Timeout: 5 * time.Second},
	}
}

// Getters / Setters

func (c *Client) GetId() int64 {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Id
}

func (c *Client) GetName() string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Name
}

func (c *Client) GetScheme() string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Scheme
}

func (c *Client) GetHost() string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Host
}

func (c *Client) GetPort() int {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Port
}

func (c *Client) Update(cfg *ClientConfig) {
	c.cfgMu.Lock()
	defer c.cfgMu.Unlock()
	c.cfg = cfg
}

// API calls

func (c *Client) getBaseURL() string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return fmt.Sprintf("%s://%s:%d/api", c.cfg.Scheme, c.cfg.Host, c.cfg.Port)
}

func (c *Client) ensureSession() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.session.ValidUntil) {
		c.logger.Debug().Msg("using existing valid Pi-hole session")
		return nil // session still valid
	}
	c.logger.Debug().Msg("requesting new Pi-hole session")

	// call POST /auth
	authResp, err := c.Login()
	if err != nil {
		c.logger.Error().Err(err).Msg("error logging via /auth endpoint")
	}

	sidPrefix := authResp.Session.SID
	if len(sidPrefix) > 8 {
		sidPrefix = sidPrefix[:8]
	}
	c.logger.Debug().Str("sid_prefix", sidPrefix).Int("validity_seconds", authResp.Session.Validity).Msg("Pi-hole session established")

	if !authResp.Session.Valid {
		c.logger.Error().Msg("invalid session")
		return fmt.Errorf("auth failed: session invalid")
	}

	c.session = sessionState{
		SID:        authResp.Session.SID,
		ValidUntil: time.Now().Add(time.Duration(authResp.Session.Validity) * time.Second),
	}

	return nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	c.logger.Debug().Str("method", req.Method).Str("url", req.URL.String()).Msg("sending request to Pi-hole")

	if err := c.ensureSession(); err != nil {
		return nil, err
	}
	req.Header.Set("X-FTL-SID", c.session.SID)
	return c.HTTP.Do(req)
}

func (c *Client) GetNodeInfo() PiholeNode {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return PiholeNode{
		Id:   c.cfg.Id,
		Host: c.cfg.Host,
	}
}

func (c *Client) FetchQueryLogs(req FetchQueryLogClientRequest) (*FetchQueryLogResponse, error) {
	query := buildQueryParams(req)
	c.logger.Debug().Str("query", query).Msg("fetching query logs from Pi-hole")

	url := fmt.Sprintf("%s/queries?%s", c.getBaseURL(), query)
	httpReq, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(httpReq)
	if err != nil {
		return nil, fmt.Errorf("requesting pihole logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var result FetchQueryLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode Pi-hole response")
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

func (c *Client) GetDomainRules(opts GetDomainRulesOptions) (*GetDomainRulesResponse, error) {
	path := "/domains"
	switch {
	case opts.Type != nil && opts.Kind != nil && opts.Domain != nil:
		path = fmt.Sprintf("/domains/%s/%s/%s", *opts.Type, *opts.Kind, url.PathEscape(*opts.Domain))
	case opts.Type != nil && opts.Kind != nil:
		path = fmt.Sprintf("/domains/%s/%s", *opts.Type, *opts.Kind)
	case opts.Type != nil && opts.Domain != nil:
		path = fmt.Sprintf("/domains/%s/%s", *opts.Type, url.PathEscape(*opts.Domain))
	case opts.Type != nil:
		path = fmt.Sprintf("/domains/%s", *opts.Type)
	case opts.Kind != nil:
		path = fmt.Sprintf("/domains/%s", *opts.Kind)
	case opts.Domain != nil:
		path = fmt.Sprintf("/domains/%s", url.PathEscape(*opts.Domain))
	}

	url := c.getBaseURL() + path
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("requesting Pi-hole domain rules: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var result GetDomainRulesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode Pi-hole response")
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

func (c *Client) AddDomainRule(opts AddDomainRuleOptions) (*AddDomainRuleResponse, error) {
	c.logger.Debug().Str("type", opts.Type).Str("kind", opts.Kind).Msg("adding domain rule")

	url := fmt.Sprintf("%s/domains/%s/%s", c.getBaseURL(), opts.Type, opts.Kind)

	bodyBytes, err := json.Marshal(opts.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("adding domain to pihole: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var result AddDomainRuleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode Pi-hole response")
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

func (c *Client) RemoveDomainRule(opts RemoveDomainRuleOptions) error {
	c.logger.Debug().Str("type", opts.Type).Str("kind", opts.Kind).Msg("adding domain rule")

	url := fmt.Sprintf("%s/domains/%s/%s/%s", c.getBaseURL(), opts.Type, opts.Kind, opts.Domain)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("requesting Pi-hole logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) Login() (*AuthResponse, error) {
	c.logger.Debug().Msg("loggin into pihole instance")
	
	c.cfgMu.RLock()
	payload := map[string]string{"password": c.cfg.Password}
	c.cfgMu.RUnlock()
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create auth request")
		return nil, fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)	
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to authenticate with pihole")
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error().Int("status", resp.StatusCode).Msg("failed to authenticate with Pi-hole")
		return nil, fmt.Errorf("auth failed, status: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode pihole response")
		return nil, fmt.Errorf("decoding auth response: %w", err)
	}

	return &authResp, err
}

func (c *Client) Logout() error {
	c.logger.Debug().Msg("logging out pihole session")
	if !time.Now().Before(c.session.ValidUntil) {
		return nil // Session is already invalid
	}
	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	_, err := c.doRequest(req)
	return err
}
