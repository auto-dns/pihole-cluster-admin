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
)

func buildQueryParams(opts FetchLogsQueryOptions) string {
	params := url.Values{}

	if opts.From != nil {
		params.Set("from", fmt.Sprintf("%d", *opts.From))
	}
	if opts.Until != nil {
		params.Set("until", fmt.Sprintf("%d", *opts.Until))
	}
	if opts.Length != nil {
		params.Set("length", fmt.Sprintf("%d", *opts.Length))
	}
	if opts.Start != nil {
		params.Set("start", fmt.Sprintf("%d", *opts.Start))
	}
	if opts.Cursor != nil {
		params.Set("cursor", fmt.Sprintf("%d", *opts.Cursor))
	}
	if opts.Domain != nil {
		params.Set("domain", *opts.Domain)
	}
	if opts.ClientIP != nil {
		params.Set("client_ip", *opts.ClientIP)
	}
	if opts.ClientName != nil {
		params.Set("client_name", *opts.ClientName)
	}
	if opts.Upstream != nil {
		params.Set("upstream", *opts.Upstream)
	}
	if opts.Type != nil {
		params.Set("type", *opts.Type)
	}
	if opts.Status != nil {
		params.Set("status", *opts.Status)
	}
	if opts.Reply != nil {
		params.Set("reply", *opts.Reply)
	}
	if opts.DNSSEC != nil {
		params.Set("dnssec", *opts.DNSSEC)
	}
	if opts.Disk != nil {
		params.Set("disk", strconv.FormatBool(*opts.Disk))
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
}

func NewClient(cfg *config.PiholeConfig) *Client {
	return &Client{
		cfg:  cfg,
		HTTP: &http.Client{Timeout: 5 * time.Second},
	}
}

func (c *Client) getBaseURL() string {
	return fmt.Sprintf("%s://%s:%d/api", c.cfg.Scheme, c.cfg.Host, c.cfg.Port)
}

func (c *Client) ensureSession() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Now().Before(c.session.ValidUntil) {
		return nil // session still valid
	}

	// call POST /auth
	payload := map[string]string{"password": c.cfg.Password}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
		return fmt.Errorf("decoding auth response: %w", err)
	}

	if !authResp.Session.Valid {
		return fmt.Errorf("auth failed: session invalid")
	}

	c.session = sessionState{
		SID:        authResp.Session.SID,
		ValidUntil: time.Now().Add(time.Duration(authResp.Session.Validity) * time.Second),
	}

	return nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	if err := c.ensureSession(); err != nil {
		return nil, err
	}
	req.Header.Set("X-FTL-SID", c.session.SID)
	return c.HTTP.Do(req)
}

func (c *Client) FetchLogs(opts FetchLogsQueryOptions) (*QueryLogResponse, error) {
	query := buildQueryParams(opts)

	url := fmt.Sprintf("%s/queries?from=%d&until=%d", c.getBaseURL(), query)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("requesting Pi-hole logs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var result QueryLogResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	result.PiholeNode = PiholeNode{
		ID:   c.cfg.ID,
		Host: c.cfg.Host,
	}

	return &result, nil
}

func (c *Client) Logout() error {
	if !time.Now().Before(c.session.ValidUntil) {
		return nil // Session is already invalid
	}
	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, _ := http.NewRequest(http.MethodDelete, url, nil)
	_, err := c.doRequest(req)
	return err
}
