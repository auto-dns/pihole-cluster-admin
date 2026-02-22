package pihole

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	logs "github.com/auto-dns/pihole-cluster-admin/internal/logger"
	"github.com/go-chi/chi/middleware"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func buildQueryParams(req queriesWireRequest) string {
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
	cfg      *ClientConfig
	HTTP     *http.Client
	session  sessionState
	mu       sync.Mutex
	loginMu  sync.Mutex // Serializes Login: only one session per node at a time
	logger   zerolog.Logger
	cfgMu    sync.RWMutex
}

type ClientConfig struct {
	Id       int64
	Name     string
	Scheme   string
	Host     string
	Port     int
	Password string
}

func NewClient(cfg *ClientConfig, logger zerolog.Logger, opts ...ClientOption) *Client {
	l := logger.With().Int64("id", cfg.Id).Logger()

	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	c := &Client{
		cfg:    cfg,
		logger: l,
		HTTP:   &http.Client{Timeout: 5 * time.Second, Transport: tr},
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
	old := *c.cfg
	cc := *cfg
	c.cfg = &cc
	c.cfgMu.Unlock()

	if old.Host != cc.Host || old.Port != cc.Port || old.Scheme != cc.Scheme || old.Password != cc.Password {
		c.mu.Lock()
		c.session = sessionState{}
		c.mu.Unlock()
		if tr, ok := c.HTTP.Transport.(*http.Transport); ok {
			tr.CloseIdleConnections()
		}
	}
}

// API calls

func (c *Client) getBaseURL() string {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return fmt.Sprintf("%s://%s:%d/api", c.cfg.Scheme, c.cfg.Host, c.cfg.Port)
}

func (c *Client) ensureSession(ctx context.Context, force bool) (string, error) {
	if force {
		c.mu.Lock()
		sid := c.session.SID
		c.session = sessionState{}
		c.mu.Unlock()
		// Free the old session slot on Pi-hole before creating a new one
		if sid != "" {
			_ = c.Logout(ctx)
		}
	}

	// Fast path: reuse existing valid session
	leeway := 5 * time.Second
	c.mu.Lock()
	sid := c.session.SID
	validUntil := c.session.ValidUntil
	c.mu.Unlock()

	if sid != "" && time.Now().Add(leeway).Before(validUntil) {
		logs.Event(ctx, c.logger).Msg("using existing valid pihole session")
		return sid, nil
	}

	// Serialize login: only one session per node (prevents session slot exhaustion)
	c.loginMu.Lock()
	defer c.loginMu.Unlock()

	// Double-check: another goroutine may have logged in while we waited
	c.mu.Lock()
	sid = c.session.SID
	validUntil = c.session.ValidUntil
	c.mu.Unlock()
	if sid != "" && time.Now().Add(leeway).Before(validUntil) {
		return sid, nil
	}

	c.logger.Debug().Msg("requesting new pihole session")
	if err := c.Login(ctx); err != nil {
		return "", fmt.Errorf("auth failed: %w", err)
	}

	c.mu.Lock()
	sid = c.session.SID
	c.mu.Unlock()
	return sid, nil
}

func (c *Client) doRequest(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if ctx == nil {
		ctx = context.TODO()
	}

	requestId := middleware.GetReqID(ctx)
	if requestId == "" {
		requestId = uuid.NewString()
	}
	childId := fmt.Sprintf("%s:n%d", requestId, c.GetId(ctx))

	logs.Event(ctx, c.logger).
		Str("method", req.Method).
		Str("url", req.URL.String()).
		Str("child_request_id", childId).
		Msg("sending request to pihole")

	sid, err := c.ensureSession(ctx, false)
	if err != nil {
		return nil, err
	}

	req.Header.Set("X-FTL-SID", sid)
	req.Header.Set("X-Request-ID", childId)
	req.Header.Set("User-Agent", "pihole-cluster-admin/6")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()

		// Re-auth
		sid, err := c.ensureSession(ctx, true)
		if err != nil {
			return nil, err
		}

		if req.GetBody != nil {
			rc, _ := req.GetBody()
			req.Body = rc
		}

		req.Header.Set("X-FTL-SID", sid)
		req.Header.Set("X-Request-ID", childId)
		req.Header.Set("User-Agent", "pihole-cluster-admin/6")
		resp, err = c.HTTP.Do(req)
		if err != nil {
			return nil, err
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		logs.Event(ctx, c.logger).
			Str("child_request_id", childId).
			Int("status", resp.StatusCode).
			Msg("pihole responded with non-success status")
	}

	return resp, nil
}

func (c *Client) GetNodeInfo(_ context.Context) domain.PiholeNodeRef {
	c.cfgMu.RLock()
	defer c.cfgMu.RUnlock()
	return domain.PiholeNodeRef{Id: c.cfg.Id, Host: c.cfg.Host, Name: c.cfg.Name}
}

// Blocking

func (c *Client) GetBlockingState(ctx context.Context) (*domain.BlockingState, error) {
	url := c.getBaseURL() + "/dns/blocking"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("requesting Pi-hole blocking status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{
			Status: resp.StatusCode,
			Body:   string(b),
		}
	}

	var result blockingWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	blockingState := blockingWireResponseToDomain(result)
	return &blockingState, nil
}

func (c *Client) SetBlockingState(ctx context.Context, blocking bool, timer *int) (*domain.BlockingState, error) {
	evt := c.logger.Debug().Bool("blocking", blocking)
	if timer != nil {
		evt = evt.Int("timer", *timer)
	}
	evt.Msg("setting blocking state")
	url := fmt.Sprintf("%s/dns/blocking", c.getBaseURL())

	body, err := json.Marshal(setBlockingWireRequest{
		Blocking: blocking,
		Timer:    timer,
	})
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("setting blocking state on pihole: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var result blockingWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode Pi-hole response")
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	blockingState := blockingWireResponseToDomain(result)
	return &blockingState, nil
}

// Query logs

func (c *Client) FetchQueryLogs(ctx context.Context, req queriesWireRequest) (*domain.QueryLogPage, error) {
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
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}

	var w queriesWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode Pi-hole response")
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	page := &domain.QueryLogPage{
		Entries:         make([]domain.QueryLogEntry, 0, len(w.Queries)),
		Cursor:          w.Cursor,
		RecordsTotal:    w.RecordsTotal,
		RecordsFiltered: w.RecordsFiltered,
		Draw:            w.Draw,
		Took:            time.Duration(max(w.Took, 0) * float64(time.Second)),
	}

	for _, e := range w.Queries {
		sec := math.Floor(e.Time)
		nsec := (e.Time - sec) * 1e9
		replyDur := time.Duration(max(e.Reply.Time, 0) * float64(time.Second))
		page.Entries = append(page.Entries, domain.QueryLogEntry{
			Id:         e.Id,
			Time:       time.Unix(int64(sec), int64(nsec)).UTC(),
			QType:      e.Type,
			Status:     e.Status,
			DNSSEC:     e.DNSSEC,
			Domain:     e.Domain,
			Upstream:   e.Upstream,
			ReplyType:  e.Reply.Type,
			ReplyTime:  replyDur,
			ClientIP:   e.Client.IP,
			ClientName: e.Client.Name,
			ListID:     e.ListID,
			EDECode:    e.EDE.Code,
			EDEText:    e.EDE.Text,
			CNAME:      e.CNAME,
		})
	}

	return page, nil
}

// Domain rules

func toDomainRule(w domainWireInfo) domain.DomainRule {
	return domain.DomainRule{
		Id:        w.Id,
		Domain:    w.Domain,
		Unicode:   w.Unicode,
		Type:      domain.RuleType(w.Type),
		Kind:      domain.RuleKind(w.Kind),
		Comment:   w.Comment,
		Groups:    w.Groups,
		Enabled:   w.Enabled,
		CreatedAt: time.Unix(w.DateAdded, 0).UTC(),
		UpdatedAt: time.Unix(w.DateModified, 0).UTC(),
	}
}
func toDomainRuleSet(w domainsWireResponse) domain.DomainRuleSet {
	out := domain.DomainRuleSet{Rules: make([]domain.DomainRule, 0, len(w.Domains))}
	for _, d := range w.Domains {
		out.Rules = append(out.Rules, toDomainRule(d))
	}
	out.Took = time.Duration(max(w.Took, 0) * float64(time.Second))
	return out
}

func (c *Client) ListDomainRules(ctx context.Context, q domain.ListDomainRulesQuery) (*domain.DomainRuleSet, error) {
	var path string
	switch {
	case q.Type == nil && q.Kind == nil && q.Domain == nil:
		path = "/domains"
	case q.Type != nil && q.Kind == nil && q.Domain == nil:
		path = "/domains/" + url.PathEscape(string(*q.Type))
	case q.Type == nil && q.Kind != nil && q.Domain == nil:
		path = "/domains/" + url.PathEscape(string(*q.Kind))
	case q.Type == nil && q.Kind == nil && q.Domain != nil:
		path = "/domains/" + url.PathEscape(*q.Domain)
	case q.Type != nil && q.Kind != nil && q.Domain == nil:
		path = "/domains/" + url.PathEscape(string(*q.Type)) + "/" + url.PathEscape(string(*q.Kind))
	case q.Type != nil && q.Kind != nil && q.Domain != nil:
		path = "/domains/" + url.PathEscape(string(*q.Type)) + "/" + url.PathEscape(string(*q.Kind)) + "/" + url.PathEscape(*q.Domain)
	default:
		path = "/domains"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getBaseURL()+path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("pihole list domains: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}

	var w domainsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		c.logger.Error().Err(err).Msg("decode domains")
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	rs := toDomainRuleSet(w)
	return &rs, nil
}

func (c *Client) AddDomainRule(ctx context.Context, cmd domain.AddDomainRulesCommand) (*domain.AddDomainRulesResult, error) {
	c.logger.Debug().Str("type", string(cmd.Type)).Str("kind", string(cmd.Kind)).Msg("adding domain rule")

	url := fmt.Sprintf("%s/domains/%s/%s", c.getBaseURL(), cmd.Type, cmd.Kind)

	wreq := addDomainsWireRequest{
		Domain:  append([]string(nil), cmd.Domains...),
		Comment: cmd.Comment,
		Groups:  append([]int(nil), cmd.Groups...),
		Enabled: cmd.Enabled,
	}
	body, err := json.Marshal(wreq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("adding domain to pihole: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var w addDomainsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		c.logger.Error().Err(err).Msg("failed to decode Pi-hole response")
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	out := domain.AddDomainRulesResult{
		Rules: make([]domain.DomainRule, 0, len(w.Domains)),
		Took:  time.Duration(max(w.Took, 0) * float64(time.Second)),
	}

	for _, d := range w.Domains {
		out.Rules = append(out.Rules, toDomainRule(d))
	}

	if w.Processed != nil {
		if n := len(w.Processed.Success); n > 0 {
			out.Processed.Success = make([]domain.DomainProcessedItem, 0, n)
			for _, s := range w.Processed.Success {
				out.Processed.Success = append(out.Processed.Success, domain.DomainProcessedItem{Item: s.Item})
			}
		}
		if n := len(w.Processed.Errors); n > 0 {
			out.Processed.Errors = make([]domain.DomainProcessedError, 0, n)
			for _, e := range w.Processed.Errors {
				out.Processed.Errors = append(out.Processed.Errors, domain.DomainProcessedError{
					Item:  e.Item,
					Error: e.Error,
				})
			}
		}
	}

	return &out, nil
}

func (c *Client) RemoveDomainRule(ctx context.Context, cmd domain.RemoveDomainRuleCommand) error {
	c.logger.Debug().Str("type", string(cmd.Type)).Str("kind", string(cmd.Kind)).Msg("removing domain rule")

	url := fmt.Sprintf("%s/domains/%s/%s/%s", c.getBaseURL(), cmd.Type, cmd.Kind, url.PathEscape(cmd.Domain))
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

// Auth

func (c *Client) Login(ctx context.Context) error {
	c.logger.Debug().Msg("logging into pihole instance")

	c.cfgMu.RLock()
	w := authWireRequest{Password: c.cfg.Password}
	c.cfgMu.RUnlock()

	body, _ := json.Marshal(w)
	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		// Pi-hole session slots full – back off before returning to avoid hammering
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Second):
		}
		return fmt.Errorf("auth failed, status: %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed, status: %d", resp.StatusCode)
	}

	var authResp authWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}
	if !authResp.Session.Valid || authResp.Session.SID == "" {
		return fmt.Errorf("auth failed: session invalid")
	}

	c.mu.Lock()
	c.session = sessionState{
		SID:        authResp.Session.SID,
		ValidUntil: time.Now().Add(time.Duration(authResp.Session.Validity) * time.Second),
	}
	c.mu.Unlock()

	return nil
}

func (c *Client) AuthStatus(ctx context.Context) (*domain.AuthStatus, error) {
	c.logger.Trace().Msg("getting client auth status")

	url := fmt.Sprintf("%s/auth", c.getBaseURL())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		c.logger.Trace().Err(err).Msg("error preparing http request")
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("getting Pi-hole auth status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var authResp authWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		c.logger.Error().Err(err).Msg("decoding auth response")
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	took := time.Duration(max(authResp.Took, 0) * float64(time.Second))
	validUntil := time.Now().Add(time.Duration(authResp.Session.Validity) * time.Second)

	return &domain.AuthStatus{
		Valid:      authResp.Session.Valid,
		Validity:   time.Duration(authResp.Session.Validity * int(time.Second)),
		ValidUntil: validUntil,
		Took:       took,
	}, nil
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
