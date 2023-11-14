package server

import (
	"os"

	"github.com/drone/envsubst"
	"github.com/mpapenbr/iracelog-wamp-router/pkg/config"
	"gopkg.in/yaml.v3"
)

type racelogAuth struct {
	realm string
	authn *ticketKeyStore
	authz *racelogAuthz
}

func newAuth(fn string) (*racelogAuth, error) {
	buf, err := os.ReadFile(fn)
	if err != nil {
		return nil, err
	}
	x, err := envsubst.EvalEnv(string(buf))
	if err != nil {
		return nil, err
	}
	authConfig := &config.MyConfig{}
	err = yaml.Unmarshal([]byte(x), authConfig)
	if err != nil {
		return nil, err
	}
	var ret racelogAuth

	roleLookup := make(map[string]*config.Role)
	for _, r := range authConfig.Roles {
		roleLookup[r.Name] = r
	}
	ret.authz = newAuthz(roleLookup)

	principalLookup := make(map[string]*config.TicketAuth)
	for _, t := range authConfig.Auth.Tickets {
		principalLookup[t.Principal] = t
	}
	ret.realm = authConfig.Realm

	ret.authn = newAuthn(principalLookup)
	return &ret, nil
}
