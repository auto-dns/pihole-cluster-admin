Wire types:
pihole/client.go

-   [x] GetBlockingState

    -   Deserializes response from pihole to private blockingWireResponse type
    -   Returns domain.BlockingState

-   [x] Update

    -   Takes cfg \*ClientConfig (public struct)
    -   type ClientConfig struct {
        Id int64
        Name string
        Scheme string
        Host string
        Port int
        Password string
        }
    -   Should it take a domain type? Or should it take a public type exposed in its own package (as it currently does)?

-   [x] FetchQueryLogs
    -   Takes private type pihole.fetchQueryLogClientRequest
    -   Returns public type pihole.FetchQueryLogResponse
    -   type fetchQueryLogClientRequest struct {
        Filters FetchQueryLogFilters
        Cursor *int
        Length *int // number of results
        Start \*int // offset
        }
    -   type FetchQueryLogFilters struct {
        From *int64 // Unix timestamp
        Until *int64 // Unix timestamp
        Domain *string // filter by domain
        ClientIP *string // filter by client IP
        ClientName *string // filter by client hostname
        Upstream *string // filter by upstream server
        Type *string // query type (A, AAAA, etc.)
        Status *string // query status (GRAVITY, FORWARDED, etc.)
        Reply *string // reply type (NODATA, NXDOMAIN, etc.)
        DNSSEC *string // DNSSEC status (SECURE, INSECURE, etc.)
        Disk \*bool // load from on-disk database
        }
    -   type FetchQueryLogResponse struct {
        Queries []DNSLogEntry `json:"queries"`
        Cursor int `json:"cursor"`
        RecordsTotal int64 `json:"recordsTotal"`
        RecordsFiltered int64 `json:"recordsFiltered"`
        Draw int64 `json:"draw"`
        Took float64 `json:"took"`
        }
    -   type DNSLogEntry struct {
        Id int64 `json:"id"`
        Time float64 `json:"time"`
        Type string `json:"type"`
        Status string `json:"status"`
        DNSSEC string `json:"dnssec"`
        Domain string `json:"domain"`
        Upstream *string `json:"upstream"`
        Reply ReplyInfo `json:"reply"`
        Client ClientInfo `json:"client"`
        ListID *int64 `json:"list_id"`
        EDE EDEInfo `json:"ede"`
        CNAME \*string `json:"cname"`
        }
    -   type ReplyInfo struct {
        Type string `json:"type"`
        Time float64 `json:"time"`
        }
    -   type ClientInfo struct {
        IP string `json:"ip"`
        Name \*string `json:"name"`
        }
    -   type EDEInfo struct {
        Code int64 `json:"code"`
        Text \*string `json:"text"`
        }
    -   All of these are defined in the pihole package alongside client.go

*   [ ] GetAllDomainRules

    -   Returns GetDomainRulesResponse (in pihole package)
    -   type GetDomainRulesResponse struct {
        Domains []DomainInfo `json:"domains"`
        Took float64 `json:"took"`
        }
    -   type DomainInfo struct {
        Domain string `json:"domain"`
        Unicode string `json:"unicode"`
        Type string `json:"type"` // "allow" or "deny"
        Kind string `json:"kind"` // "exact" or "regex"
        Comment \*string `json:"comment,omitempty"`
        Groups []int `json:"groups"`
        Enabled bool `json:"enabled"`
        Id int `json:"id"`
        DateAdded int64 `json:"date_added"`
        DateModified int64 `json:"date_modified"`
        }

*   [ ] AddDomainRule

    -   Takes AddDomainRuleOptions (from pihole package)
    -   Returns AddDomainRuleResponse (from pihole package)
    -   type AddDomainRuleOptions struct {
        Type RuleType
        Kind RuleKind
        Payload AddDomainPayload // request body
        }
    -   type RuleType string
    -   const (
        RuleTypeAllow RuleType = "allow"
        RuleTypeDeny RuleType = "deny"
        )
    -   type RuleKind string
    -   const (
        RuleKindExact RuleKind = "exact"
        RuleKindRegex RuleKind = "regex"
        )
    -   type AddDomainRuleResponse struct {
        Domains []DomainInfo `json:"domains"`
        Processed \*ProcessedResult `json:"processed,omitempty"`
        Took float64 `json:"took"`
        }
    -   type DomainInfo struct {
        Domain string `json:"domain"`
        Unicode string `json:"unicode"`
        Type string `json:"type"` // "allow" or "deny"
        Kind string `json:"kind"` // "exact" or "regex"
        Comment \*string `json:"comment,omitempty"`
        Groups []int `json:"groups"`
        Enabled bool `json:"enabled"`
        Id int `json:"id"`
        DateAdded int64 `json:"date_added"`
        DateModified int64 `json:"date_modified"`
        }
    -   type ProcessedResult struct {
        Success []struct {
        Item string `json:"item"`
        } `json:"success"`
        Errors []struct {
        Item string `json:"item"`
        Error string `json:"error"`
        } `json:"errors"`
        }

*   [ ] RemoveDomainRule

-   Takes RemoveDomainRuleOptions (from pihole package)
-   Returns error
-   type RemoveDomainRuleOptions struct {
    Type RuleType
    Kind RuleKind
    Domain string // a single domain to remove
    }
-   type RuleType string
-   const (
    RuleTypeAllow RuleType = "allow"
    RuleTypeDeny RuleType = "deny"
    )
-   type RuleKind string
-   const (
    RuleKindExact RuleKind = "exact"
    RuleKindRegex RuleKind = "regex"
    )

*   [ ] AuthStatus

-   Returns domain.AuthStatus.
-   Deserializes pihole client response to private authResponse struct
-   type authResponse struct {
    Session struct {
    Valid bool `json:"valid"`
    SID string `json:"sid"`
    CSRF string `json:"csrf"`
    Validity int `json:"validity"`
    } `json:"session"`
    Took float64 `json:"took"`
    }

store/initstatus.go

-   [ ] GetInitializationStatus
-   Returns domain.InitStatus
-   Deserializes from db to initStatusRow (private)
-   type initStatusRow struct {
    UserCreated bool
    PiholeStatus domain.PiholeStatus
    }
-   type PiholeStatus string
-   const (
    PiholeUninitialized PiholeStatus = "UNINITIALIZED"
    PiholeAdded PiholeStatus = "ADDED"
    PiholeSkipped PiholeStatus = "SKIPPED"
    )
-   func (s *PiholeStatus) Scan(src any) error {
    switch v := src.(type) {
    case string:
    *s = PiholeStatus(v)
    case []byte:
    *s = PiholeStatus(string(v))
    default:
    return fmt.Errorf("pihole status: unsupported scan type %T", src)
    }
    if !s.IsValid() {
    return fmt.Errorf("pihole status: invalid value %q", string(*s))
    }
    return nil
    }
-   type InitStatus struct {
    UserCreated bool `json:"userCreated"`
    PiholeStatus PiholeStatus `json:"piholeStatus"`
    }

-   [ ] SetPiholeStatus
-   takes domain.PiholeStatus type, writes it to db directly as string
-   type PiholeStatus string
-   const (
    PiholeUninitialized PiholeStatus = "UNINITIALIZED"
    PiholeAdded PiholeStatus = "ADDED"
    PiholeSkipped PiholeStatus = "SKIPPED"
    )
-   // driver.Valuer implementation for writing directly to DB
    func (s PiholeStatus) Value() (driver.Value, error) {
    if !s.IsValid() {
    return nil, fmt.Errorf("pihole status: invalid value %q", string(s))
    }
    return string(s), nil
    }
-   type InitStatus struct {
    UserCreated bool `json:"userCreated"`
    PiholeStatus PiholeStatus `json:"piholeStatus"`
    }

store/pihole.go

-   [x] AddPiholeNode
-   takes public AddPiholeParams
-   Returns domain.PiholeNode
-   type AddPiholeParams struct {
    Scheme string
    Host string
    Port int
    Name string
    Description string
    Password string
    }

-   [x] UpdatePiholeNode
-   takes publick UpdatePiholeParams
-   Returns domain.PiholeNode
-   type UpdatePiholeParams struct {
    Scheme *string
    Host *string
    Port *int
    Name *string
    Description *string
    Password *string
    }

-   [ ] GetPiholeNode and GetAllPiholeNodes
-   Returns domain.PiholeNode
-   Gets value from db as private piholeRow struct, deserializes to domain.PiholeNode using a private helper function
-   type piholeRow struct {
    Id int64
    Scheme string
    Host string
    Port int
    Name string
    Description string
    PasswordEnc string
    CreatedAt time.Time
    UpdatedAt time.Time
    }

-   GetPiholeNodeSecret
-   Returns domain.PiholeNodeSecret
-   Gets value from db as private piholeRow struct, deserializes to domain.PiholeNodeSecret using a private helper function
-   type piholeRow struct {
    Id int64
    Scheme string
    Host string
    Port int
    Name string
    Description string
    PasswordEnc string
    CreatedAt time.Time
    UpdatedAt time.Time
    }

store/session.go

-   [ ] CreateSession
-   Takes CreateSessionParams (in store package)
-   Returns domain.Session
-   type CreateSessionParams struct {
    Id string
    UserId int64
    ExpiresAt time.Time
    }

-   [ ] GetAllSessions, GetSession
-   Returns domain.Session type
-   Deserializes from db to private sessionRow type, converts to domain.Session using helper function
-   type sessionRow struct {
    Id string
    UserId int64
    CreatedAt time.Time
    ExpiresAt time.Time
    }

store/user.go

-   [ ] CreateUser
-   Takes CreateUserParams (defined in store package)
-   Returns domain.User
-   type CreateUserParams struct {
    Username string
    Password string
    }

-   [ ] GetUser
-   Returns domain.User
-   Deserializes from db to private userRow and converts to domain.User using helper function
-   type userRow struct {
    Id int64
    Username string
    PasswordHash string
    CreatedAt time.Time
    UpdatedAt time.Time
    }

-   [ ] GetUserAuth
-   Returns domain.UserAuth
-   Deserializes from db to private userRow and converts to domain.UserAuth using helper function
-   type userRow struct {
    Id int64
    Username string
    PasswordHash string
    CreatedAt time.Time
    UpdatedAt time.Time
    }

-   [ ] ValidateUser
-   Returns domain.UserAuth
-   Deserializes from db to private userRow and converts to domain.User using helper function
-   type userRow struct {
    Id int64
    Username string
    PasswordHash string
    CreatedAt time.Time
    UpdatedAt time.Time
    }

-   [ ] UpdateUser
-   Takes UpdateUserParams defined in store
-   Returns domain.User
-   type UpdateUserParams struct {
    Username *string
    Password *string
    }

Domain types:
cluster.go
service

DTO types:
handler

<!-- Responses -->

You’ve got the right idea: adapters (Pi-hole + SQLite) should decode their own wire shapes and return domain types; only the HTTP layer should deal in DTOs. Below is a concrete plan that you can apply across the app, plus per-function guidance based on your inventory.

Golden rules (apply everywhere)

wire → domain happens inside the adapter (pihole, store).

Keep all wire structs private (unexported) in the adapter.

Convert to domain types before returning.

service consumes/returns domain only (plus service “snapshots” also made of domain primitives).

handler maps domain → DTO (the JSON you expose). No domain mutations here, just shaping.

Prefer time.Time/time.Duration in domain; convert from/into numbers/strings at the edges (wire/DTO).

For commands (create/update requests), let each adapter define its own params types, or define domain commands and write adapter mappers—pick one approach and be consistent.

Where “wire” lives

Pi-hole adapter (internal/pihole): \*\_wire.go files (unexported structs) for HTTP JSON.

Store adapter (internal/store): private row structs and scanners are your “wire” to SQLite—keep them unexported and convert to domain in the store.

You already do this in store (good!). Just finish doing it in pihole by making the public response structs private wire structs and returning domain instead.

Function-by-function recommendations
pihole/client.go
✅ GetBlockingState

Current: decodes into blockingWireResponse (private), returns domain.BlockingState.

Keep as-is. Good.

🔧 Update (client config)

Goal: keep adapters decoupled.

Keep pihole.ClientConfig as an adapter config type. Your service can build it from domain.PiholeNodeSecret (fetched from the store).

Optionally add a helper:

func NewClientFromNode(n domain.PiholeNodeSecret, logger zerolog.Logger) \*Client

Internally maps domain → ClientConfig. But don’t make the adapter accept domain everywhere; keep adapter config local to the adapter.

🔁 FetchQueryLogs

Current: takes fetchQueryLogClientRequest (private) and returns public pihole.FetchQueryLogResponse with DNSLogEntry etc.

Change to:

Keep all Pi-hole JSON shapes private:

// pihole/queries*wire.go
type queriesWireRequest struct { /* wire filters, cursor, etc. \_/ }
type queriesWireResponse struct {
Queries []queryWireEntry `json:"queries"`
Cursor int `json:"cursor"`
RecordsTotal int64 `json:"recordsTotal"`
RecordsFiltered int64 `json:"recordsFiltered"`
Draw int64 `json:"draw"`
Took float64 `json:"took"`
}
type queryWireEntry struct {
ID int64 `json:"id"`
Time float64 `json:"time"` // seconds.fraction (if you later confirm ms, change conversion)
Type string `json:"type"`
Status string `json:"status"`
DNSSEC string `json:"dnssec"`
Domain string `json:"domain"`
Upstream *string `json:"upstream"`
Reply struct {
Type string `json:"type"`
Time float64 `json:"time"` // seconds.fraction
} `json:"reply"`
Client struct {
IP string `json:"ip"`
Name *string `json:"name"`
} `json:"client"`
ListID *int64 `json:"list_id"`
EDE struct {
Code int64 `json:"code"`
Text *string `json:"text"`
} `json:"ede"`
CNAME \*string `json:"cname"`
}

Convert to domain:

// domain/logs.go
type QueryLogEntry struct {
ID int64
Time time.Time
QType string
Status string
DNSSEC string
Domain string
Upstream *string
ReplyType string
ReplyTime time.Duration
ClientIP string
ClientName *string
ListID *int64
EDECode int64
EDEText *string
CNAME \*string
}
type QueryLogPage struct {
Entries []QueryLogEntry
Cursor int
RecordsTotal int64
RecordsFiltered int64
Draw int64
Took time.Duration
}

In the client, convert float64 seconds → time.Time / time.Duration. (If you discover the API uses milliseconds, switch the conversion in one place; domain stays stable.)

Return: (\*domain.QueryLogPage, error).

🔁 GetAllDomainRules

Current: returns public GetDomainRulesResponse.

Change to:

Make the wire struct private:

type domainsWireResponse struct {
Domains []domainWireInfo `json:"domains"`
Took float64 `json:"took"`
}
type domainWireInfo struct {
Domain string `json:"domain"`
Unicode string `json:"unicode"`
Type string `json:"type"` // "allow"|"deny"
Kind string `json:"kind"` // "exact"|"regex"
Comment \*string `json:"comment,omitempty"`
Groups []int `json:"groups"`
Enabled bool `json:"enabled"`
ID int `json:"id"`
DateAdded int64 `json:"date_added"` // unix seconds
DateModified int64 `json:"date_modified"` // unix seconds
}

Convert to domain:

// domain/rules.go
type RuleType string
const ( RuleTypeAllow RuleType = "allow"; RuleTypeDeny RuleType = "deny" )
type RuleKind string
const ( RuleKindExact RuleKind = "exact"; RuleKindRegex RuleKind = "regex" )

type DomainRule struct {
ID int
Domain string
Unicode string
Type RuleType
Kind RuleKind
Comment \*string
Groups []int
Enabled bool
DateAdded time.Time
DateModified time.Time
}

type DomainRulesResult struct {
Domains []DomainRule
Took time.Duration
}

Return: (\*domain.DomainRulesResult, error).

🔁 AddDomainRule

Current: takes AddDomainRuleOptions and returns AddDomainRuleResponse from pihole package.

Change to:

Define a domain command and result:

// domain/rules.go
type AddDomainRuleCommand struct {
Type RuleType
Kind RuleKind
Domains []string
Comment *string
Groups []int
Enabled *bool
}
type AddDomainRuleResult struct {
Domains []DomainRule
Processed struct {
Success []string
Errors []struct {
Item string
Error string
}
}
Took time.Duration
}

In the client, map the domain command → wire payload Pi-hole expects, and map the wire response → AddDomainRuleResult.

Return: (\*domain.AddDomainRuleResult, error).

🔁 RemoveDomainRule

Current: takes RemoveDomainRuleOptions, returns error.

Keep: Accept a domain command:

type RemoveDomainRuleCommand struct {
Type RuleType
Kind RuleKind
Domain string
}

Map to wire in client; keep return error.

✅ AuthStatus

Already returns domain.AuthStatus from a private authResponse. Good.

store layer (SQLite adapter)

You’re already doing the right thing: private row structs and scanners; return domain types.

store/initstatus.go

GetInitializationStatus: returns domain.InitStatus built from initStatusRow. Good.

SetPiholeStatus: accepts domain.PiholeStatus. Good.

store/pihole.go

Add/Update/Get/GetAll:

Params (AddPiholeParams, UpdatePiholeParams) can stay in store as adapter command types.

Return domain (domain.PiholeNode, domain.PiholeNodeSecret). Good.

Alternative: define domain commands (e.g., domain.CreatePiholeNode) and keep tiny mappers in the store. Pick one approach and use it everywhere. Since you already have store-scoped params, keep them and map in the service.

store/session.go, store/user.go

Keep returning domain (domain.Session, domain.User, domain.UserAuth) and keep row structs private. Good.

TL;DR mapping per layer

Adapter (pihole/store)

Input: adapter params (config/filters/IDs).

Decode wire (HTTP JSON / SQL rows) → private wire/row types.

Convert to domain.

Return domain (+ typed errors as needed).

Service

Accept domain inputs (or simple scalars) from handlers.

Call adapters.

Aggregate into service snapshots made of domain primitives (e.g., ClusterBlockingSnapshot).

Return domain/service to handlers.

Handler

Map domain/service → DTO (JSON).

For POST/PUT, map DTO → domain command and pass to service.

No business logic here.

Suggested renames / file layout (makes it obvious)
internal/
pihole/
client.go
cluster.go
blocking_wire.go
queries_wire.go
domains_wire.go
auth_wire.go
store/
initstatus.go (private rows here)
pihole.go (private rows here)
session.go (private rows here)
user.go (private rows here)
domain/
blocking.go
logs.go
rules.go
pihole_nodes.go
result.go (NodeResult[T] + ErrorMessage() helper)
service/
clusterblocking/
service.go
types.go (snapshot/summary types)
handler/
clusterblocking/
handler.go
dto.go (HTTP JSON-only shapes)

Final notes on a couple of your bullets

Client.Update(cfg \*ClientConfig): keep it as an adapter config. The service can fetch domain.PiholeNodeSecret from the store and map into ClientConfig. This avoids coupling the adapter to your persistence model and keeps the adapter easy to reuse or test.

For all your current public structs in the pihole package that mirror wire JSON (FetchQueryLogResponse, DNSLogEntry, GetDomainRulesResponse, etc.): make them private wire structs and return domain equivalents instead. That’s the main cleanup.

If you want, I can draft one full conversion (e.g., FetchQueryLogs: wire → domain) as a reference implementation you can replicate.
