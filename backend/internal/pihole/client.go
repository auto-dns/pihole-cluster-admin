package pihole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

type ClientOption func(*Client)

func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		if hc != nil {
			c.HTTP = hc
		}
	}
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

func NewClient(cfg *ClientConfig, logger zerolog.Logger, opts ...ClientOption) ClientInterface {
	c := &Client{
		cfg:    cfg,
		logger: logger,
		HTTP:   &http.Client{Timeout: 5 * time.Second},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// Getters / Setters

func (c *Client) GetId(_ context.Context) int64 {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Id
}

func (c *Client) GetName(_ context.Context) string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Name
}

func (c *Client) GetScheme(_ context.Context) string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Scheme
}

func (c *Client) GetHost(_ context.Context) string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Host
}

func (c *Client) GetPort(_ context.Context) int {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return c.cfg.Port
}

func (c *Client) Update(_ context.Context, cfg *ClientConfig) {
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

func (c *Client) ensureSession(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Refresh slightly before session expiry
	leeway := 5 * time.Second
	if time.Now().Add(leeway).Before(c.session.ValidUntil) {
		c.logger.Debug().Msg("using existing valid pihole session")
		return c.session.SID, nil
	}
	
	c.logger.Debug().Msg("requesting new pihole session")
	authResp, err := c.Login(ctx)
	if err != nil {
		return "", fmt.Errorf("auth failed: %w", err)
	}
	if !authResp.Session.Valid {
		return "", fmt.Errorf("auth failed: session invalid")
	}

	c.session = sessionState{
		SID:        authResp.Session.SID,
		ValidUntil: time.Now().Add(time.Duration(authResp.Session.Validity) * time.Second),
	}

	sidPrefix := c.session.SID
	if len(sidPrefix) > 8 {
		sidPrefix = sidPrefix[:8]
	}
	c.logger.Debug().
		Str("sid_prefix", sidPrefix).
		Int("validity_seconds", authResp.Session.Validity).
		Msg("Pi-hole session established")

	return c.session.SID, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	c.logger.Debug().Str("method", req.Method).Str("url", req.URL.String()).Msg("sending request to pihole")

	ctx := req.Context()
	if ctx == nil {
        ctx = context.TODO()
    }

	sid, err := c.ensureSession(ctx)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-FTL-SID", sid)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()

		// Re-auth
		sid, err := c.ensureSession(ctx)
		if err != nil {
			return nil, err
		}
		
		if req.GetBody != nil {
    		rc, _ := req.GetBody()
    		req.Body = rc
		}

		req.Header.Set("X-FTL-SID", sid)
		return c.HTTP.Do(req)
	}

	return resp, nil
}

func (c *Client) GetNodeInfo(_ context.Context) PiholeNode {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return PiholeNode{
		Id:   c.cfg.Id,
		Host: c.cfg.Host,
	}
}

func (c *Client) FetchQueryLogs(ctx context.Context, req FetchQueryLogClientRequest) (*FetchQueryLogResponse, error) {
	query := buildQueryParams(req)
	c.logger.Debug().Str("query", query).Msg("fetching query logs from Pi-hole")

	url := fmt.Sprintf("%s/queries?%s", c.getBaseURL(), query)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func (c *Client) GetDomainRules(ctx context.Context, opts GetDomainRulesOptions) (*GetDomainRulesResponse, error) {
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
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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

func (c *Client) AddDomainRule(ctx context.Context, opts AddDomainRuleOptions) (*AddDomainRuleResponse, error) {
	c.logger.Debug().Str("type", opts.Type).Str("kind", opts.Kind).Msg("adding domain rule")

	url := fmt.Sprintf("%s/domains/%s/%s", c.getBaseURL(), opts.Type, opts.Kind)

	bodyBytes, err := json.Marshal(opts.Payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.GetBody = func() (io.ReadCloser, error) {
    	return io.NopCloser(bytes.NewReader(bodyBytes)), nil
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

func (c *Client) RemoveDomainRule(ctx context.Context, opts RemoveDomainRuleOptions) error {
	c.logger.Debug().Str("type", opts.Type).Str("kind", opts.Kind).Msg("removing domain rule")

	url := fmt.Sprintf("%s/domains/%s/%s/%s", c.getBaseURL(), opts.Type, opts.Kind, url.PathEscape(opts.Domain))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
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

func (c *Client) Login(ctx context.Context) (*AuthResponse, error) {
	c.logger.Debug().Msg("logging into pihole instance")
	
	c.cfgMu.RLock()
	payload := map[string]string{"password": c.cfg.Password}
	c.cfgMu.RUnlock()

	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)	
	if err != nil {
		return nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth failed, status: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return nil, fmt.Errorf("decoding auth response: %w", err)
	}

	return &authResp, nil
}

func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	sid := c.session.SID
	c.session = sessionState{}
	c.mu.Unlock()

	if sid == "" {
		return nil
	}

	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("creating logout request: %w", err)
	}
	req.Header.Set("X-FTL-SID", sid)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("logout request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
        c.logger.Warn().Int("status", resp.StatusCode).Msg("unexpected status code on logout")
    }

	return nil
}
