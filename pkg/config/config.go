package config

var (
	WSAddr       string // listen addr for websocket
	TCPAddr      string // listen addr for raw TCP
	RouterConfig string // router configuration
	LogLevel     string // sets the log level (zap log level values)
	LogFormat    string // text vs json
)

type MyConfig struct {
	Realm string  `json:"realm"`
	Roles []*Role `json:"roles"`
	Auth  Auth    `json:"auth" yaml:"auth"`
}

type Role struct {
	Name        string
	Permissions []*Permission `json:"permissions" yaml:"permissions"`
}

type Permission struct {
	URI   string      `json:"uri" yaml:"uri"`
	Match string      `json:"match" yaml:"match"`
	Allow AllowAction `json:"allow" yaml:"allow"`
}

type AllowAction struct {
	Call      bool `json:"call" yaml:"call"`
	Register  bool `json:"register" yaml:"register"`
	Publish   bool `json:"publish" yaml:"publish"`
	Subscribe bool `json:"subscribe" yaml:"subscribe"`
}

type Auth struct {
	Tickets []*TicketAuth `json:"tickets" yaml:"tickets"`
}
type TicketAuth struct {
	Principal string `json:"principal" yaml:"principal"`
	Ticket    string `json:"ticket"`
	Role      string `json:"role"`
}
