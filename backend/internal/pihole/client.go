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

func cloneValues(v url.Values) url.Values {
	out := make(url.Values, len(v))
	for k, vals := range v {
		out[k] = append([]string(nil), vals...)
	}
	return out
}

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
		HTTP:   &http.Client{Timeout: 15 * time.Second, Transport: tr},
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
		// Free the old session slot on Pi-hole before creating a new one.
		// Must call logoutWithSID directly — Logout() re-reads c.session which is
		// already cleared above and would short-circuit without sending the DELETE.
		if sid != "" {
			_ = c.logoutWithSID(ctx, sid)
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

const (
	retryMax      = 3
	retryBaseMs   = 250
)

// doWithRetry executes an HTTP request, retrying on network-level errors
// (e.g. connection refused during a transient Pi-hole restart) with
// exponential backoff. It never retries on HTTP-level errors to preserve
// idempotency for mutating requests.
func (c *Client) doWithRetry(req *http.Request) (*http.Response, error) {
	var (
		resp *http.Response
		err  error
	)
	for attempt := 0; attempt <= retryMax; attempt++ {
		if attempt > 0 {
			delay := time.Duration(retryBaseMs*math.Pow(2, float64(attempt-1))) * time.Millisecond
			select {
			case <-req.Context().Done():
				return nil, req.Context().Err()
			case <-time.After(delay):
			}
			// Rewind body for retry if possible
			if req.GetBody != nil {
				rc, berr := req.GetBody()
				if berr != nil {
					return nil, berr
				}
				req.Body = rc
			}
			c.logger.Debug().Int("attempt", attempt+1).Str("url", req.URL.String()).Msg("retrying pihole request after network error")
		}
		resp, err = c.HTTP.Do(req)
		if err == nil {
			return resp, nil
		}
	}
	return nil, err
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

	resp, err := c.doWithRetry(req)
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
		resp, err = c.doWithRetry(req)
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

// Stats

func (c *Client) GetStatsSummary(ctx context.Context) (*domain.StatsSummary, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getBaseURL()+"/stats/summary", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("requesting stats summary: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w statsSummaryWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	return &domain.StatsSummary{
		QueriesTotal:   w.Queries.Total,
		QueriesBlocked: w.Queries.Blocked,
		BlockedPercent: w.Queries.PercentBlocked,
		GravitySize:    w.Gravity.DomainsBeingBlocked,
		UniqueClients:  w.Clients.Active,
		UniqueDomains:  w.Queries.UniqueDomains,
	}, nil
}

func (c *Client) GetStatsHistory(ctx context.Context, from, until *int64) (*domain.StatsHistory, error) {
	params := url.Values{}
	if from != nil {
		params.Set("from", strconv.FormatInt(*from, 10))
	}
	if until != nil {
		params.Set("until", strconv.FormatInt(*until, 10))
	}
	u := c.getBaseURL() + "/history/database"
	if q := params.Encode(); q != "" {
		u += "?" + q
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("requesting stats history: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w statsHistoryWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	h := &domain.StatsHistory{
		Points: make([]domain.StatsHistoryPoint, 0, len(w.History)),
	}
	for _, e := range w.History {
		h.Points = append(h.Points, domain.StatsHistoryPoint{
			Timestamp: time.Unix(e.Timestamp, 0).UTC(),
			Total:     e.Total,
			Blocked:   e.Blocked,
		})
	}
	return h, nil
}

func (c *Client) GetStatsTopDomains(ctx context.Context, from, until *int64, count *int) (*domain.StatsTopDomains, error) {
	params := url.Values{}
	if from != nil {
		params.Set("from", strconv.FormatInt(*from, 10))
	}
	if until != nil {
		params.Set("until", strconv.FormatInt(*until, 10))
	}
	if count != nil {
		params.Set("count", strconv.Itoa(*count))
	}
	base := c.getBaseURL() + "/stats/database/top_domains"

	// Two calls: queried (default) then blocked (?blocked=true) — Pi-hole returns
	// only one list per request, toggled by the blocked query param.
	queriedURL := base
	if q := params.Encode(); q != "" {
		queriedURL += "?" + q
	}
	queried, err := c.fetchTopDomains(ctx, queriedURL)
	if err != nil {
		return nil, fmt.Errorf("fetching queried domains: %w", err)
	}
	blockedParams := cloneValues(params)
	blockedParams.Set("blocked", "true")
	blocked, err := c.fetchTopDomains(ctx, base+"?"+blockedParams.Encode())
	if err != nil {
		return nil, fmt.Errorf("fetching blocked domains: %w", err)
	}
	return &domain.StatsTopDomains{TopQueried: queried, TopBlocked: blocked}, nil
}

func (c *Client) fetchTopDomains(ctx context.Context, u string) ([]domain.TopDomainEntry, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("requesting top domains: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w statsTopDomainsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	out := make([]domain.TopDomainEntry, 0, len(w.Domains))
	for _, d := range w.Domains {
		out = append(out, domain.TopDomainEntry{Domain: d.Domain, Count: d.Count})
	}
	return out, nil
}

func (c *Client) GetStatsTopClients(ctx context.Context, from, until *int64, count *int) (*domain.StatsTopClients, error) {
	params := url.Values{}
	if from != nil {
		params.Set("from", strconv.FormatInt(*from, 10))
	}
	if until != nil {
		params.Set("until", strconv.FormatInt(*until, 10))
	}
	if count != nil {
		params.Set("count", strconv.Itoa(*count))
	}
	u := c.getBaseURL() + "/stats/database/top_clients"
	if q := params.Encode(); q != "" {
		u += "?" + q
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("requesting top clients: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w statsTopClientsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	out := &domain.StatsTopClients{
		Clients: make([]domain.TopClientEntry, 0, len(w.Clients)),
	}
	for _, s := range w.Clients {
		out.Clients = append(out.Clients, domain.TopClientEntry{IP: s.IP, Name: s.Name, Count: s.Count})
	}
	return out, nil
}

// Adlists

// unixOrZero converts a Unix timestamp to UTC time, returning the zero time.Time
// when ts is 0 (Pi-hole uses 0 to mean "never", e.g. date_updated on a new list).
func unixOrZero(ts int64) time.Time {
	if ts == 0 {
		return time.Time{}
	}
	return time.Unix(ts, 0).UTC()
}

func toAdlist(w adlistWireEntry) domain.Adlist {
	return domain.Adlist{
		Id:             w.Id,
		Address:        w.Address,
		Type:           domain.AdlistType(w.Type),
		Comment:        w.Comment,
		Groups:         w.Groups,
		Enabled:        w.Enabled,
		DateAdded:      time.Unix(w.DateAdded, 0).UTC(),
		DateModified:   time.Unix(w.DateModified, 0).UTC(),
		DateUpdated:    unixOrZero(w.DateUpdated),
		Number:         w.Number,
		InvalidDomains: w.InvalidDomains,
		Status:         w.Status,
	}
}

func (c *Client) ListAdlists(ctx context.Context) (*domain.AdlistSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getBaseURL()+"/lists", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("listing adlists: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w listsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	out := &domain.AdlistSet{Lists: make([]domain.Adlist, 0, len(w.Lists))}
	for _, e := range w.Lists {
		out.Lists = append(out.Lists, toAdlist(e))
	}
	return out, nil
}

func (c *Client) AddAdlist(ctx context.Context, cmd domain.AddAdlistCommand) (*domain.AdlistSet, error) {
	wreq := addAdlistWireRequest{
		Address: cmd.Address,
		Type:    string(cmd.Type),
		Comment: cmd.Comment,
		Groups:  cmd.Groups,
		Enabled: cmd.Enabled,
	}
	body, err := json.Marshal(wreq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.getBaseURL()+"/lists", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("adding adlist: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w listsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	out := &domain.AdlistSet{Lists: make([]domain.Adlist, 0, len(w.Lists))}
	for _, e := range w.Lists {
		out.Lists = append(out.Lists, toAdlist(e))
	}
	return out, nil
}

func (c *Client) UpdateAdlist(ctx context.Context, cmd domain.UpdateAdlistCommand) (*domain.AdlistSet, error) {
	wreq := updateAdlistWireRequest{
		Enabled: cmd.Enabled,
		Comment: cmd.Comment,
		Groups:  cmd.Groups,
	}
	body, err := json.Marshal(wreq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	u := fmt.Sprintf("%s/lists/%d", c.getBaseURL(), cmd.Id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("updating adlist: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w listsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	out := &domain.AdlistSet{Lists: make([]domain.Adlist, 0, len(w.Lists))}
	for _, e := range w.Lists {
		out.Lists = append(out.Lists, toAdlist(e))
	}
	return out, nil
}

func (c *Client) RemoveAdlist(ctx context.Context, cmd domain.RemoveAdlistCommand) error {
	u := fmt.Sprintf("%s/lists/%d", c.getBaseURL(), cmd.Id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("removing adlist: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	return nil
}

// RebuildGravity calls POST /api/action/gravity. The endpoint streams text/plain
// output (pihole -g progress) and always commits HTTP 200 before the work begins,
// so the response body must be drained before returning.
func (c *Client) RebuildGravity(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.getBaseURL()+"/action/gravity", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("rebuilding gravity: %w", err)
	}
	defer resp.Body.Close()
	// Drain the streamed output so the connection can be reused.
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &httpStatusError{Status: resp.StatusCode}
	}
	return nil
}

func (c *Client) RestartDNS(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.getBaseURL()+"/action/restartDNS", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("restarting DNS: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &httpStatusError{Status: resp.StatusCode}
	}
	return nil
}

func (c *Client) GetVersionInfo(ctx context.Context) (*domain.NodeVersionInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getBaseURL()+"/info/version", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("getting version: %w", err)
	}
	defer resp.Body.Close()
	var w versionWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding version: %w", err)
	}
	return &domain.NodeVersionInfo{
		PiholeVersion: w.Version.Core.Local.Version,
		FTLVersion:    w.Version.FTL.Local.Version,
	}, nil
}

func (c *Client) GetDatabaseInfo(ctx context.Context) (*domain.NodeDBInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getBaseURL()+"/stats/summary", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("getting gravity info: %w", err)
	}
	defer resp.Body.Close()
	var w gravityInfoWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding gravity info: %w", err)
	}
	info := &domain.NodeDBInfo{GravityCount: w.Gravity.DomainsBeingBlocked}
	if w.Gravity.LastUpdate > 0 {
		t := time.Unix(w.Gravity.LastUpdate, 0)
		info.GravityUpdatedAt = &t
	}
	return info, nil
}

func (c *Client) FlushCache(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.getBaseURL()+"/action/flush/cache", nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("flushing cache: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK {
		return &httpStatusError{Status: resp.StatusCode}
	}
	return nil
}

func (c *Client) SetPassword(ctx context.Context, newPassword string) error {
	var w setPasswordWireRequest
	w.Webserver.API.Password = newPassword
	body, err := json.Marshal(w)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.getBaseURL()+"/config", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("setting password: %w", err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return &httpStatusError{Status: resp.StatusCode}
	}
	return nil
}

// Groups

func toGroup(w groupWireEntry) domain.Group {
	return domain.Group{
		Id:           w.Id,
		Name:         w.Name,
		Description:  w.Description,
		Enabled:      w.Enabled,
		DateAdded:    time.Unix(w.DateAdded, 0).UTC(),
		DateModified: time.Unix(w.DateModified, 0).UTC(),
	}
}

func toGroupSet(w groupsWireResponse) domain.GroupSet {
	out := domain.GroupSet{Groups: make([]domain.Group, 0, len(w.Groups))}
	for _, g := range w.Groups {
		out.Groups = append(out.Groups, toGroup(g))
	}
	return out
}

func (c *Client) ListGroups(ctx context.Context) (*domain.GroupSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getBaseURL()+"/groups", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("listing groups: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w groupsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	gs := toGroupSet(w)
	return &gs, nil
}

func (c *Client) AddGroup(ctx context.Context, cmd domain.AddGroupCommand) (*domain.GroupSet, error) {
	wreq := addGroupWireRequest{
		Name:        cmd.Name,
		Description: cmd.Description,
		Enabled:     cmd.Enabled,
	}
	body, err := json.Marshal(wreq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.getBaseURL()+"/groups", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("adding group: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w groupsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	gs := toGroupSet(w)
	return &gs, nil
}

func (c *Client) UpdateGroup(ctx context.Context, cmd domain.UpdateGroupCommand) (*domain.GroupSet, error) {
	wreq := updateGroupWireRequest{
		Description: cmd.Description,
		Enabled:     cmd.Enabled,
	}
	body, err := json.Marshal(wreq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	u := fmt.Sprintf("%s/groups/%s", c.getBaseURL(), url.PathEscape(cmd.Name))
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("updating group: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w groupsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	gs := toGroupSet(w)
	return &gs, nil
}

func (c *Client) RemoveGroup(ctx context.Context, cmd domain.RemoveGroupCommand) error {
	u := fmt.Sprintf("%s/groups/%s", c.getBaseURL(), url.PathEscape(cmd.Name))
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("removing group: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	return nil
}

// Clients

func toClient(w clientWireEntry) domain.PiholeClient {
	groups := w.Groups
	if groups == nil {
		groups = []int{}
	}
	return domain.PiholeClient{
		Id:           w.Id,
		IP:           w.IP,
		Name:         w.Name,
		Comment:      w.Comment,
		Groups:       groups,
		DateAdded:    time.Unix(w.DateAdded, 0).UTC(),
		DateModified: time.Unix(w.DateModified, 0).UTC(),
	}
}

func toClientSet(w clientsWireResponse) domain.PiholeClientSet {
	out := domain.PiholeClientSet{Clients: make([]domain.PiholeClient, 0, len(w.Clients))}
	for _, c := range w.Clients {
		out.Clients = append(out.Clients, toClient(c))
	}
	return out
}

func (c *Client) ListPiholeClients(ctx context.Context) (*domain.PiholeClientSet, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.getBaseURL()+"/clients", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("listing clients: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w clientsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	cs := toClientSet(w)
	return &cs, nil
}

func (c *Client) UpdatePiholeClient(ctx context.Context, cmd domain.UpdatePiholeClientCommand) (*domain.PiholeClientSet, error) {
	groups := cmd.Groups
	if groups == nil {
		groups = []int{}
	}
	wreq := updateClientWireRequest{
		Groups:  groups,
		Comment: cmd.Comment,
	}
	body, err := json.Marshal(wreq)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	u := fmt.Sprintf("%s/clients/%d", c.getBaseURL(), cmd.Id)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, u, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("updating client: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var w clientsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&w); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	cs := toClientSet(w)
	return &cs, nil
}

func (c *Client) RemovePiholeClient(ctx context.Context, cmd domain.RemovePiholeClientCommand) error {
	u := fmt.Sprintf("%s/clients/%d", c.getBaseURL(), cmd.Id)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return fmt.Errorf("removing client: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	return nil
}

func (c *Client) TestRegex(ctx context.Context, testDomain string) (*domain.RegexTestResult, error) {
	body, err := json.Marshal(regexTestWireRequest{Domain: testDomain})
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.getBaseURL()+"/regex/test", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	req.GetBody = func() (io.ReadCloser, error) {
		return io.NopCloser(bytes.NewReader(body)), nil
	}
	resp, err := c.doRequest(req)
	if err != nil {
		return nil, fmt.Errorf("testing regex: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, &httpStatusError{Status: resp.StatusCode, Body: string(b)}
	}
	var wireResp domainsWireResponse
	if err := json.NewDecoder(resp.Body).Decode(&wireResp); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}
	result := &domain.RegexTestResult{Domain: testDomain}
	for _, d := range wireResp.Domains {
		result.Matches = append(result.Matches, domain.RegexMatch{
			ID:      d.Id,
			Pattern: d.Domain,
			Kind:    d.Type,
			Enabled: d.Enabled,
		})
	}
	return result, nil
}

func (c *Client) logoutWithSID(ctx context.Context, sid string) error {
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

func (c *Client) Logout(ctx context.Context) error {
	c.mu.Lock()
	sid := c.session.SID
	c.session = sessionState{}
	c.mu.Unlock()
	return c.logoutWithSID(ctx, sid)
}
