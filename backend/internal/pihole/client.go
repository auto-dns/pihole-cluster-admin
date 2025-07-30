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

	"github.com/auto-dns/pihole-cluster-admin/internal/config"
	"github.com/rs/zerolog"
)

func buildQueryParams(req FetchQueryLogRequest) string {
	params := url.Values{}

	// Pagination
	if req.CursorID != nil {
		params.Set("cursor", *req.CursorID)
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
	cfg     *config.PiholeConfig
	HTTP    *http.Client
	session sessionState
	mu      sync.Mutex
	logger  zerolog.Logger
}

func NewClient(cfg *config.PiholeConfig, logger zerolog.Logger) *Client {
	return &Client{
		cfg:    cfg,
		logger: logger,
		HTTP:   &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) getBaseURL() string {
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
	payload := map[string]string{"password": c.cfg.Password}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to create auth request")
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		c.logger.Error().Err(err).Msg("failed to authenticate with pihole")
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error().Int("status", resp.StatusCode).Msg("failed to authenticate with Pi-hole")
		return fmt.Errorf("auth failed, status: %d", resp.StatusCode)
	}

	var authResp struct {
		Session struct {
			Valid    bool   `json:"valid"`
			SID      string `json:"sid"`
			CSRF     string `json:"csrf"`
			Validity int    `json:"validity"`
		} `json:"session"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode Pi-hole response")
		return fmt.Errorf("decoding auth response: %w", err)
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
	return PiholeNode{
		ID:   c.cfg.ID,
		Host: c.cfg.Host,
	}
}

func (c *Client) FetchQueryLogs(req FetchQueryLogRequest) (*FetchQueryLogResponse, error) {
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

func (c *Client) Logout() error {
	c.logger.Debug().Msg("logging out Pi-hole session")
	if !time.Now().Before(c.session.ValidUntil) {
		return nil // Session is already invalid
	}
	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	_, err := c.doRequest(req)
	return err
}
