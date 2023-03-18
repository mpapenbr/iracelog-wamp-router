package server

import (
	"github.com/gammazero/nexus/v3/wamp"

	"github.com/mpapenbr/iracelog-wamp-router/log"
	"github.com/mpapenbr/iracelog-wamp-router/pkg/config"
)

type racelogAuthz struct {
	roleLookup map[string]*config.Role
}

func newAuthz(roleLookup map[string]*config.Role) *racelogAuthz {
	return &racelogAuthz{roleLookup: roleLookup}
}

//nolint:lll //keeping one line seems better
func (a *racelogAuthz) Authorize(sess *wamp.Session, msg wamp.Message) (bool, error) {
	authrole := wamp.OptionString(sess.Details, "authrole")
	var r *config.Role
	var ok bool
	if r, ok = a.roleLookup[authrole]; !ok {
		return false, nil
	}
	log.Debug("Authorize role", log.String("authrole", authrole))
	switch m := msg.(type) {
	case *wamp.Subscribe:
		return checkAccess(m.Topic, r, func(p *config.Permission) bool { return p.Allow.Subscribe })
	case *wamp.Publish:
		return checkAccess(m.Topic, r, func(p *config.Permission) bool { return p.Allow.Publish })
	case *wamp.Call:
		return checkAccess(m.Procedure, r, func(p *config.Permission) bool { return p.Allow.Call })
	case *wamp.Register:
		return checkAccess(m.Procedure, r, func(p *config.Permission) bool { return p.Allow.Register })
	case *wamp.Unregister, *wamp.Unsubscribe, *wamp.Yield, *wamp.Goodbye:
		return true, nil
	}
	return false, nil
}

//nolint:whitespace //can't make both linter and editor happy
func checkAccess(
	uRI wamp.URI, role *config.Role, checkFlag func(p *config.Permission) bool,
) (bool, error) {
	log.Debug("checking access", log.String("uri", string(uRI)))
	for i := range role.Permissions {
		p := role.Permissions[i]
		switch p.Match {
		case "prefix":
			if uRI.PrefixMatch(wamp.URI(p.URI)) && checkFlag(p) {
				return true, nil
			}
		case "match":
			if uRI.WildcardMatch(wamp.URI(p.URI)) && checkFlag(p) {
				return true, nil
			}
		}
	}
	log.Debug("Access denied", log.String("uri", string(uRI)))
	return false, nil
}
