package auth

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"
)

const defaultHTTPTimeout = 5 * time.Second

type CognitoVerifier struct {
	issuer   string
	audience string
	jwksURL  string
	client   *http.Client
	now      func() time.Time
}

type cognitoOption func(*CognitoVerifier)

func WithHTTPClient(client *http.Client) cognitoOption {
	return func(verifier *CognitoVerifier) {
		verifier.client = client
	}
}

func WithJWKSURL(jwksURL string) cognitoOption {
	return func(verifier *CognitoVerifier) {
		verifier.jwksURL = jwksURL
	}
}

func WithNow(now func() time.Time) cognitoOption {
	return func(verifier *CognitoVerifier) {
		verifier.now = now
	}
}

func NewCognitoVerifier(issuer string, audience string, options ...cognitoOption) (*CognitoVerifier, error) {
	issuer = strings.TrimRight(strings.TrimSpace(issuer), "/")
	audience = strings.TrimSpace(audience)
	if issuer == "" {
		return nil, errors.New("APP_COGNITO_ISSUER is required")
	}
	if audience == "" {
		return nil, errors.New("APP_COGNITO_AUDIENCE is required")
	}

	verifier := &CognitoVerifier{
		issuer:   issuer,
		audience: audience,
		jwksURL:  issuer + "/.well-known/jwks.json",
		client:   &http.Client{Timeout: defaultHTTPTimeout},
		now:      time.Now,
	}
	for _, option := range options {
		option(verifier)
	}
	if verifier.client == nil {
		verifier.client = &http.Client{Timeout: defaultHTTPTimeout}
	}
	if verifier.now == nil {
		verifier.now = time.Now
	}
	return verifier, nil
}

func (verifier *CognitoVerifier) VerifyRequest(r *http.Request) error {
	const bearerPrefix = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, bearerPrefix) {
		return ErrUnauthorized
	}
	return verifier.VerifyToken(r.Context(), strings.TrimSpace(strings.TrimPrefix(header, bearerPrefix)))
}

func (verifier *CognitoVerifier) VerifyToken(ctx context.Context, token string) error {
	if verifier == nil || token == "" {
		return ErrUnauthorized
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ErrUnauthorized
	}

	var header jwtHeader
	if err := decodeJWTPart(parts[0], &header); err != nil {
		return ErrUnauthorized
	}
	if header.Algorithm != "RS256" || header.KeyID == "" {
		return ErrUnauthorized
	}

	var claims jwtClaims
	if err := decodeJWTPart(parts[1], &claims); err != nil {
		return ErrUnauthorized
	}
	if err := verifier.validateClaims(claims); err != nil {
		return ErrUnauthorized
	}

	key, err := verifier.findPublicKey(ctx, header.KeyID)
	if err != nil {
		return ErrUnauthorized
	}
	signingInput := parts[0] + "." + parts[1]
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return ErrUnauthorized
	}
	digest := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(key, crypto.SHA256, digest[:], signature); err != nil {
		return ErrUnauthorized
	}

	return nil
}

func (verifier *CognitoVerifier) validateClaims(claims jwtClaims) error {
	if claims.Issuer != verifier.issuer {
		return fmt.Errorf("issuer mismatch")
	}
	if claims.ExpiresAt <= verifier.now().Unix() {
		return fmt.Errorf("token expired")
	}
	if claims.TokenUse != "" && claims.TokenUse != "access" {
		return fmt.Errorf("token_use must be access")
	}
	if claims.ClientID == verifier.audience {
		return nil
	}
	for _, audience := range claims.Audience.values() {
		if audience == verifier.audience {
			return nil
		}
	}
	return fmt.Errorf("audience mismatch")
}

func (verifier *CognitoVerifier) findPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, verifier.jwksURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := verifier.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jwks status: %d", response.StatusCode)
	}

	var keySet jwks
	if err := json.NewDecoder(response.Body).Decode(&keySet); err != nil {
		return nil, err
	}
	for _, key := range keySet.Keys {
		if key.KeyID != keyID || key.KeyType != "RSA" {
			continue
		}
		return key.publicKey()
	}
	return nil, fmt.Errorf("key not found: %s", keyID)
}

func decodeJWTPart(part string, destination any) error {
	decoded, err := base64.RawURLEncoding.DecodeString(part)
	if err != nil {
		return err
	}
	return json.Unmarshal(decoded, destination)
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	KeyID     string `json:"kid"`
}

type jwtClaims struct {
	Issuer    string      `json:"iss"`
	ExpiresAt int64       `json:"exp"`
	ClientID  string      `json:"client_id"`
	TokenUse  string      `json:"token_use"`
	Audience  audienceSet `json:"aud"`
}

type audienceSet []string

func (audience *audienceSet) UnmarshalJSON(data []byte) error {
	var one string
	if err := json.Unmarshal(data, &one); err == nil {
		*audience = []string{one}
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return err
	}
	*audience = many
	return nil
}

func (audience audienceSet) values() []string {
	return []string(audience)
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

type jwk struct {
	KeyType  string `json:"kty"`
	KeyID    string `json:"kid"`
	Use      string `json:"use"`
	Modulus  string `json:"n"`
	Exponent string `json:"e"`
}

func (key jwk) publicKey() (*rsa.PublicKey, error) {
	modulus, err := base64.RawURLEncoding.DecodeString(key.Modulus)
	if err != nil {
		return nil, err
	}
	exponent, err := base64.RawURLEncoding.DecodeString(key.Exponent)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(modulus)
	e := new(big.Int).SetBytes(exponent).Int64()
	if n.Sign() <= 0 || e <= 0 {
		return nil, fmt.Errorf("invalid rsa key")
	}
	return &rsa.PublicKey{N: n, E: int(e)}, nil
}
