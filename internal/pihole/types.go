package pihole

import "context"

// This file holds the package's operation contract: one Input and one Output
// struct per exported call, named after the call itself. Every network-touching
// call carries a Ctx field; nil means context.Background().

// NewClientInput configures a Client. Insecure skips TLS verification, which is
// needed for self-signed Pi-hole certificates.
type NewClientInput struct {
	BaseURL  string
	Password string
	Insecure bool
}

// RequestInput is one raw Pi-hole API call. Body is JSON-marshalled when non-nil.
type RequestInput struct {
	Ctx    context.Context // nil means context.Background()
	Method string
	Path   string
	Body   any
}

// RequestOutput carries the unparsed response so callers can check status
// before decoding.
type RequestOutput struct {
	StatusCode int
	Body       []byte
}

// --- groups ---

// ListGroupsInput takes no data; Ctx is the only field.
type ListGroupsInput struct {
	Ctx context.Context // nil means context.Background()
}

type ListGroupsOutput struct {
	Groups []Group
}

// GetGroupInput looks up a single group by name. Group is nil when not found.
type GetGroupInput struct {
	Ctx  context.Context // nil means context.Background()
	Name string
}

type GetGroupOutput struct {
	Group *Group
}

type CreateGroupInput struct {
	Ctx   context.Context // nil means context.Background()
	Group Group
}

type CreateGroupOutput struct {
	Group *Group
}

type UpdateGroupInput struct {
	Ctx   context.Context // nil means context.Background()
	Name  string
	Group Group
}

type DeleteGroupInput struct {
	Ctx  context.Context // nil means context.Background()
	Name string
}

type GroupIDByNameInput struct {
	Ctx  context.Context // nil means context.Background()
	Name string
}

type GroupIDByNameOutput struct {
	ID int
}

type GroupNameByIDInput struct {
	Ctx context.Context // nil means context.Background()
	ID  int
}

type GroupNameByIDOutput struct {
	Name string
}

type ResolveGroupNamesInput struct {
	Ctx   context.Context // nil means context.Background()
	Names []string
}

type ResolveGroupNamesOutput struct {
	IDs []int
}

type ResolveGroupIDsInput struct {
	Ctx context.Context // nil means context.Background()
	IDs []int
}

type ResolveGroupIDsOutput struct {
	Names []string
}

// --- clients ---

type ListClientsInput struct {
	Ctx context.Context // nil means context.Background()
}

type ListClientsOutput struct {
	Clients []ClientEntry
}

// GetClientInput looks up a client record by address. Entry is nil when the
// client has no explicit record (it then falls back to the Default group).
type GetClientInput struct {
	Ctx  context.Context // nil means context.Background()
	Addr string
}

type GetClientOutput struct {
	Entry *ClientEntry
}

type CreateClientInput struct {
	Ctx   context.Context // nil means context.Background()
	Entry ClientEntry
}

type CreateClientOutput struct {
	Entry *ClientEntry
}

type UpdateClientInput struct {
	Ctx   context.Context // nil means context.Background()
	Addr  string
	Entry ClientEntry
}

type DeleteClientInput struct {
	Ctx  context.Context // nil means context.Background()
	Addr string
}

// --- domains ---

type GetDomainInput struct {
	Ctx    context.Context // nil means context.Background()
	Domain string
	Type   string
	Kind   string
}

// GetDomainOutput carries the matched entry, or a nil Domain when none exists.
type GetDomainOutput struct {
	Domain *Domain
}

type CreateDomainInput struct {
	Ctx    context.Context // nil means context.Background()
	Domain Domain
}

type CreateDomainOutput struct {
	Domain *Domain
}

type UpdateDomainInput struct {
	Ctx    context.Context // nil means context.Background()
	ID     int
	Domain Domain
}

type DeleteDomainInput struct {
	Ctx    context.Context // nil means context.Background()
	Type   string
	Kind   string
	Domain string
}

// --- lists ---

type GetListByAddressInput struct {
	Ctx     context.Context // nil means context.Background()
	Address string
}

// GetListByAddressOutput carries the matched list, or nil when none exists.
type GetListByAddressOutput struct {
	List *List
}

type CreateListInput struct {
	Ctx  context.Context // nil means context.Background()
	List List
}

type CreateListOutput struct {
	List *List
}

type UpdateListInput struct {
	Ctx  context.Context // nil means context.Background()
	List List
}

type DeleteListInput struct {
	Ctx     context.Context // nil means context.Background()
	Address string
	Type    string
}

type TriggerGravityInput struct {
	Ctx context.Context // nil means context.Background()
}

// --- DNS records ---

type ListDNSRecordsInput struct {
	Ctx context.Context // nil means context.Background()
}

type ListDNSRecordsOutput struct {
	Records []DNSRecord
}

// GetDNSRecordInput looks up a record by its exact IP/hostname pair.
type GetDNSRecordInput struct {
	Ctx      context.Context // nil means context.Background()
	IP       string
	Hostname string
}

type GetDNSRecordOutput struct {
	Record *DNSRecord
}

type CreateDNSRecordInput struct {
	Ctx    context.Context // nil means context.Background()
	Record DNSRecord
}

type DeleteDNSRecordInput struct {
	Ctx    context.Context // nil means context.Background()
	Record DNSRecord
}

// --- CNAME records ---

type ListCNAMERecordsInput struct {
	Ctx context.Context // nil means context.Background()
}

type ListCNAMERecordsOutput struct {
	Records []CNAMERecord
}

type GetCNAMERecordInput struct {
	Ctx    context.Context // nil means context.Background()
	Domain string
}

type GetCNAMERecordOutput struct {
	Record *CNAMERecord
}

type CreateCNAMERecordInput struct {
	Ctx    context.Context // nil means context.Background()
	Record CNAMERecord
}

type DeleteCNAMERecordInput struct {
	Ctx    context.Context // nil means context.Background()
	Record CNAMERecord
}

// --- server config ---

type GetServerConfigInput struct {
	Ctx context.Context // nil means context.Background()
}

type GetServerConfigOutput struct {
	Config *ServerConfig
}

type UpdateServerConfigInput struct {
	Ctx    context.Context // nil means context.Background()
	Config *ServerConfig
}

// ctxOrBackground substitutes context.Background() for a nil Ctx, so call sites
// never need an explicit background context.
func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
