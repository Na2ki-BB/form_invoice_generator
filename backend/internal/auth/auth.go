package auth

import (
	"errors"
	"net"
	"net/http"
	"os"
	"strings"
)

type Mode string

const (
	ModeLocal          Mode = "local"
	ModeCognito        Mode = "cognito"
	ModeTrustedGateway Mode = "trusted_gateway"
)

var ErrUnauthorized = errors.New("admin authentication required")

type Authenticator struct {
	mode     Mode
	verifier *CognitoVerifier
}

func NewLocalAuthenticator() *Authenticator {
	return &Authenticator{mode: ModeLocal}
}

func NewFromEnv() (*Authenticator, error) {
	mode := Mode(strings.TrimSpace(os.Getenv("APP_AUTH_MODE")))
	if mode == "" {
		mode = ModeLocal
	}

	switch mode {
	case ModeLocal:
		return &Authenticator{mode: mode}, nil
	case ModeTrustedGateway:
		return &Authenticator{mode: mode}, nil
	case ModeCognito:
		issuer := strings.TrimSpace(os.Getenv("APP_COGNITO_ISSUER"))
		audience := strings.TrimSpace(os.Getenv("APP_COGNITO_AUDIENCE"))
		verifier, err := NewCognitoVerifier(issuer, audience)
		if err != nil {
			return nil, err
		}
		return &Authenticator{mode: mode, verifier: verifier}, nil
	default:
		return nil, errors.New("APP_AUTH_MODE must be local, cognito, or trusted_gateway")
	}
}

func (authenticator *Authenticator) AuthenticateAdmin(r *http.Request) error {
	if authenticator == nil {
		return ErrUnauthorized
	}

	switch authenticator.mode {
	case ModeLocal:
		return authenticateLocalAdmin(r)
	case ModeTrustedGateway:
		return nil
	case ModeCognito:
		return authenticator.verifier.VerifyRequest(r)
	default:
		return ErrUnauthorized
	}
}

func authenticateLocalAdmin(r *http.Request) error {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return ErrUnauthorized
	}
	ip := net.ParseIP(host)
	if ip == nil || !ip.IsLoopback() || r.Header.Get("X-Local-Admin") != "true" {
		return ErrUnauthorized
	}
	return nil
}
