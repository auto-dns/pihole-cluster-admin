package v1

import (
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"time"

	"github.com/auto-dns/pihole-cluster-admin/internal/domain"
	"github.com/auto-dns/pihole-cluster-admin/internal/errs"
)

type queryLogRequestDTO struct {
	Cursor     *string
	Length     *int
	Start      *int
	From       *time.Time
	Until      *time.Time
	Domain     *string
	ClientIP   *string
	ClientName *string
	Upstream   *string
	Type       *string
	Status     *string
	Reply      *string
	DNSSEC     *string
	Disk       *bool
}

func parseQueryLogParams(q url.Values) (queryLogRequestDTO, error) {
	var req queryLogRequestDTO

	// cursor-mode
	if cursor := q.Get("cursor"); cursor != "" {
		req.Cursor = &cursor
		if v := q.Get("length"); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return queryLogRequestDTO{}, errs.Invalid("invalid length", errors.New("invalid length"))
			}
			req.Length = &n
		}
		return req, nil
	}

	// filter-mode defaults
	until := time.Now().UTC()
	from := until.Add(-5 * time.Minute)
	req.From = &from
	req.Until = &until

	parseInt := func(k, label string) (*int, error) {
		if v := q.Get(k); v != "" {
			n, err := strconv.Atoi(v)
			if err != nil || n < 0 {
				return nil, errs.Invalid(fmt.Sprintf("invalid %s", label), fmt.Errorf("invalid %s: %w", label, err))
			}
			return &n, nil
		}
		return nil, nil
	}
	if n, err := parseInt("length", "length"); err != nil {
		return queryLogRequestDTO{}, errs.Invalid("error parsing length", err)
	} else {
		req.Length = n
	}
	if n, err := parseInt("start", "start"); err != nil {
		return queryLogRequestDTO{}, errs.Invalid("error parsing start", err)
	} else {
		req.Start = n
	}

	parseTime := func(k, label string) (*time.Time, error) {
		if v := q.Get(k); v != "" {
			t, err := time.Parse(time.RFC3339, v)
			if err != nil {
				return nil, errs.Invalid(fmt.Sprintf("invalid '%s' time", label), fmt.Errorf("invalid '%s' time: %w", label, err))
			}
			return &t, nil
		}
		return nil, nil
	}
	if t, err := parseTime("from", "from"); err != nil {
		return queryLogRequestDTO{}, err
	} else if t != nil {
		req.From = t
	}
	if t, err := parseTime("until", "until"); err != nil {
		return queryLogRequestDTO{}, err
	} else if t != nil {
		req.Until = t
	}

	setStr := func(k string) *string {
		if v := q.Get(k); v != "" {
			vv := v
			return &vv
		}
		return nil
	}
	req.Domain = setStr("domain")
	req.ClientIP = setStr("client_ip")
	req.ClientName = setStr("client_name")
	req.Upstream = setStr("upstream")
	req.Type = setStr("type")
	req.Status = setStr("status")
	req.Reply = setStr("reply")
	req.DNSSEC = setStr("dnssec")

	if v := q.Get("disk"); v != "" {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return queryLogRequestDTO{}, errs.Invalid("invalid disk", err)
		}
		req.Disk = &b
	}

	return req, nil
}

func (d queryLogRequestDTO) queryLogReqDTOToDomain() domain.QueryLogQuery {
	return domain.QueryLogQuery{
		Cursor: d.Cursor,
		Length: d.Length,
		Start:  d.Start,
		Filters: domain.QueryLogFilters{
			From:       d.From,
			Until:      d.Until,
			Domain:     d.Domain,
			ClientIP:   d.ClientIP,
			ClientName: d.ClientName,
			Upstream:   d.Upstream,
			Type:       d.Type,
			Status:     d.Status,
			Reply:      d.Reply,
			DNSSEC:     d.DNSSEC,
			Disk:       d.Disk,
		},
	}
}

type queryLogResponseDTO struct {
	Cursor       string            `json:"cursor"`
	EndOfResults bool              `json:"endOfResults"`
	Nodes        []queryLogNodeDTO `json:"nodes"`
}

type queryLogNodeDTO struct {
	Node    piholeNodeRefDTO `json:"node"`
	Success bool             `json:"success"`
	Error   string           `json:"error,omitempty"`
	Page    *queryLogPageDTO `json:"page,omitempty"`
}

type queryLogPageDTO struct {
	Cursor          int                `json:"cursor"`
	RecordsTotal    int64              `json:"recordsTotal"`
	RecordsFiltered int64              `json:"recordsFiltered"`
	Draw            int64              `json:"draw"`
	TookMs          int64              `json:"tookMs"`
	Entries         []queryLogEntryDTO `json:"entries"`
}

type queryLogEntryDTO struct {
	Id          int64   `json:"id"`
	Time        string  `json:"time"` // RFC3339 UTC
	QType       string  `json:"qtype"`
	Status      string  `json:"status"`
	DNSSEC      string  `json:"dnssec"`
	Domain      string  `json:"domain"`
	Upstream    *string `json:"upstream,omitempty"`
	ReplyType   string  `json:"replyType"`
	ReplyTimeMs int64   `json:"replyTimeMs"`
	ClientIP    string  `json:"clientIp"`
	ClientName  *string `json:"clientName,omitempty"`
	ListID      *int64  `json:"listId,omitempty"`
	EDECode     int64   `json:"edeCode"`
	EDEText     *string `json:"edeText,omitempty"`
	CNAME       *string `json:"cname,omitempty"`
}

func queryLogResponseFromDomain(res *domain.ClusterQueryLogResponse) queryLogResponseDTO {
	out := queryLogResponseDTO{
		Cursor:       res.Cursor,
		EndOfResults: res.EndOfResults,
		Nodes:        make([]queryLogNodeDTO, 0, len(res.Results)),
	}

	ids := make([]int64, 0, len(res.Results))
	for id := range res.Results {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	for _, id := range ids {
		nr := res.Results[id]
		node := queryLogNodeDTO{
			Node: piholeNodeRefDTO{
				Id:   nr.PiholeNode.Id,
				Name: nr.PiholeNode.Name,
				Host: nr.PiholeNode.Host,
			},
			Success: nr.Success,
			Error:   nr.ErrorMessage(),
		}
		if nr.Success && nr.Response != nil {
			node.Page = toPageDTO(nr.Response)
		}
		out.Nodes = append(out.Nodes, node)
	}

	return out
}

func toPageDTO(p *domain.QueryLogPage) *queryLogPageDTO {
	if p == nil {
		return nil
	}
	dto := &queryLogPageDTO{
		Cursor:          p.Cursor,
		RecordsTotal:    p.RecordsTotal,
		RecordsFiltered: p.RecordsFiltered,
		Draw:            p.Draw,
		TookMs:          p.Took.Milliseconds(),
		Entries:         make([]queryLogEntryDTO, 0, len(p.Entries)),
	}
	for _, e := range p.Entries {
		dto.Entries = append(dto.Entries, queryLogEntryDTO{
			Id:          e.Id,
			Time:        rfc3339(e.Time),
			QType:       e.QType,
			Status:      e.Status,
			DNSSEC:      e.DNSSEC,
			Domain:      e.Domain,
			Upstream:    e.Upstream,
			ReplyType:   e.ReplyType,
			ReplyTimeMs: e.ReplyTime.Milliseconds(),
			ClientIP:    e.ClientIP,
			ClientName:  e.ClientName,
			ListID:      e.ListID,
			EDECode:     e.EDECode,
			EDEText:     e.EDEText,
			CNAME:       e.CNAME,
		})
	}
	return dto
}

func rfc3339(t time.Time) string {
	return t.UTC().Format(time.RFC3339)
}
