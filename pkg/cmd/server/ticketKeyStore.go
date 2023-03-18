package server

import (
	"errors"

	"github.com/mpapenbr/iracelog-wamp-router/log"
	"github.com/mpapenbr/iracelog-wamp-router/pkg/config"
)

type ticketKeyStore struct {
	principalLookup map[string]*config.TicketAuth
}

func newAuthn(principalLookup map[string]*config.TicketAuth) *ticketKeyStore {
	return &ticketKeyStore{principalLookup: principalLookup}
}

func (ks *ticketKeyStore) Provider() string {
	return "StaticYaml"
}

func (ks *ticketKeyStore) AuthKey(authid, authmethod string) ([]byte, error) {
	log.Debug("requesting",
		log.String("authid", authid), log.String("authmethod", authmethod))
	if authmethod != "ticket" {
		return nil, errors.New("invalid authmethod")
	}
	if t, ok := ks.principalLookup[authid]; ok {
		return []byte(t.Ticket), nil
	}
	return []byte(""), nil
}

//nolint:whitespace //can't make both editor and linter happy
func (ks *ticketKeyStore) PasswordInfo(authid string) (
	salt string, keylen, iterations int,
) {
	return "", 0, 0 // not used in this keyStore
}

func (ks *ticketKeyStore) AuthRole(authid string) (string, error) {
	if v, ok := ks.principalLookup[authid]; ok {
		return v.Role, nil
	}
	return "", nil
}
