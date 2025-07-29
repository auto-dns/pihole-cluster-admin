package pihole

// DNSLogEntry models one entry in the Pi-hole query log response.
type DNSLogEntry struct {
	ID       int64      `json:"id"`
	Time     float64    `json:"time"`
	Type     string     `json:"type"`
	Status   string     `json:"status"`
	DNSSEC   string     `json:"dnssec"`
	Domain   string     `json:"domain"`
	Upstream *string    `json:"upstream"`
	Reply    ReplyInfo  `json:"reply"`
	Client   ClientInfo `json:"client"`
	ListID   *int64     `json:"list_id"`
	EDE      EDEInfo    `json:"ede"`
	CNAME    *string    `json:"cname"`
}

type ReplyInfo struct {
	Type string  `json:"type"`
	Time float64 `json:"time"`
}

type ClientInfo struct {
	IP   string  `json:"ip"`
	Name *string `json:"name"`
}

type EDEInfo struct {
	Code int64   `json:"code"`
	Text *string `json:"text"`
}

type PiholeNode struct {
	ID   string `json:"id"`
	Host string `json:"host"`
}

// QueryLogResponse models the full response
type QueryLogResponse struct {
	Queries         []DNSLogEntry `json:"queries"`
	Cursor          int64         `json:"cursor"`
	RecordsTotal    int64         `json:"recordsTotal"`
	RecordsFiltered int64         `json:"recordsFiltered"`
	Draw            int64         `json:"draw"`
	Took            float64       `json:"took"`
	PiholeNode      PiholeNode    `json:"piholeNode"`
}
