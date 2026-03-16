// Package cplnauth provides a custom BuildKit session auth provider that
// delegates token fetching to the client for registries that don't support
// the standard Docker OAuth2 token exchange (POST with grant_type=password).
//
// When --cache-auth-basic is set and --cache-ref targets a matching registry,
// this provider wraps the default Docker auth provider. It returns a real
// ed25519 public key from GetTokenAuthority, which causes the BuildKit daemon
// to call the client's FetchToken RPC instead of performing its own HTTP token
// exchange. The client's FetchToken then calls the registry's token endpoint
// using HTTP GET with Basic auth, which is the V2 token auth spec's
// "direct authentication" flow and works with registries (like Control Plane)
// that reject the OAuth2 POST flow.
package cplnauth

import (
	"context"
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth"
	"golang.org/x/crypto/nacl/sign"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// AuthProvider wraps the default Docker auth provider and overrides
// FetchToken for registries in the basicAuthHosts set. For those hosts,
// it performs a V2 token auth GET with Basic credentials instead of
// relying on the daemon's OAuth2 POST flow.
type AuthProvider struct {
	inner          session.Attachable
	config         *configfile.ConfigFile
	basicAuthHosts map[string]bool
	seed           []byte
}

// New creates a new AuthProvider. basicAuthHosts is the set of registry
// hosts (e.g. "myorg.registry.cpln.io") that should use client-side
// token fetching with Basic auth.
func New(inner session.Attachable, cfg *configfile.ConfigFile, basicAuthHosts []string) *AuthProvider {
	hosts := make(map[string]bool, len(basicAuthHosts))
	for _, h := range basicAuthHosts {
		hosts[h] = true
	}

	// Deterministic seed derived from the hosts — stable across calls
	// but unique enough for ed25519 key derivation.
	mac := hmac.New(sha256.New, []byte("cpln-auth-provider"))
	for _, h := range basicAuthHosts {
		mac.Write([]byte(h))
	}

	return &AuthProvider{
		inner:          inner,
		config:         cfg,
		basicAuthHosts: hosts,
		seed:           mac.Sum(nil),
	}
}

func (ap *AuthProvider) Register(server *grpc.Server) {
	auth.RegisterAuthServer(server, ap)
}

// Credentials delegates to the inner provider.
func (ap *AuthProvider) Credentials(ctx context.Context, req *auth.CredentialsRequest) (*auth.CredentialsResponse, error) {
	if inner, ok := ap.inner.(auth.AuthServer); ok {
		return inner.Credentials(ctx, req)
	}
	return &auth.CredentialsResponse{}, nil
}

// FetchToken handles token fetching. For basic-auth hosts, it calls the
// registry's token endpoint directly using HTTP GET with Basic auth (the V2
// "direct authentication" flow). For other hosts, it delegates to the inner
// provider.
func (ap *AuthProvider) FetchToken(ctx context.Context, req *auth.FetchTokenRequest) (*auth.FetchTokenResponse, error) {
	if !ap.basicAuthHosts[req.Host] {
		if inner, ok := ap.inner.(auth.AuthServer); ok {
			return inner.FetchToken(ctx, req)
		}
		return &auth.FetchTokenResponse{}, nil
	}

	log.Debugf("[cpln-auth] fetching token for %s via Basic auth GET", req.Host)

	// Get credentials from docker config.
	ac, err := ap.config.GetAuthConfig(req.Host)
	if err != nil {
		return nil, fmt.Errorf("cpln-auth: get auth config for %s: %w", req.Host, err)
	}

	if ac.Username == "" || ac.Password == "" {
		return nil, fmt.Errorf("cpln-auth: no credentials found for %s in docker config", req.Host)
	}

	// Build the V2 token auth GET request.
	tokenURL, err := url.Parse(req.Realm)
	if err != nil {
		return nil, fmt.Errorf("cpln-auth: parse realm %q: %w", req.Realm, err)
	}

	params := tokenURL.Query()
	if req.Service != "" {
		params.Set("service", req.Service)
	}
	for _, scope := range req.Scopes {
		params.Add("scope", scope)
	}
	tokenURL.RawQuery = params.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("cpln-auth: create token request: %w", err)
	}
	httpReq.SetBasicAuth(ac.Username, ac.Password)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cpln-auth: token request to %s failed: %w", req.Realm, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, fmt.Errorf("cpln-auth: token endpoint %s returned %d", req.Realm, resp.StatusCode)
	}

	var tokenResp struct {
		Token       string    `json:"token"`
		AccessToken string    `json:"access_token"`
		ExpiresIn   int64     `json:"expires_in"`
		IssuedAt    time.Time `json:"issued_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("cpln-auth: decode token response: %w", err)
	}

	token := tokenResp.Token
	if token == "" {
		token = tokenResp.AccessToken
	}
	if token == "" {
		return nil, fmt.Errorf("cpln-auth: empty token from %s", req.Realm)
	}

	log.Debugf("[cpln-auth] successfully obtained token for %s (expires_in=%d)", req.Host, tokenResp.ExpiresIn)

	expiresIn := tokenResp.ExpiresIn
	if expiresIn == 0 {
		expiresIn = 60
	}

	fetchResp := &auth.FetchTokenResponse{
		Token:     token,
		ExpiresIn: expiresIn,
	}
	if !tokenResp.IssuedAt.IsZero() {
		fetchResp.IssuedAt = tokenResp.IssuedAt.Unix()
	}
	return fetchResp, nil
}

// GetTokenAuthority returns an ed25519 public key for basic-auth hosts.
// This signals to the BuildKit daemon that it should delegate token fetching
// to the client (via FetchToken RPC) instead of doing its own HTTP token
// exchange.
func (ap *AuthProvider) GetTokenAuthority(ctx context.Context, req *auth.GetTokenAuthorityRequest) (*auth.GetTokenAuthorityResponse, error) {
	if !ap.basicAuthHosts[req.Host] {
		if inner, ok := ap.inner.(auth.AuthServer); ok {
			return inner.GetTokenAuthority(ctx, req)
		}
		return nil, status.Errorf(codes.Unavailable, "no token authority for %s", req.Host)
	}

	key := ap.deriveKey(req.Host, req.Salt)
	return &auth.GetTokenAuthorityResponse{PublicKey: key[32:]}, nil
}

// VerifyTokenAuthority verifies the token authority for basic-auth hosts.
func (ap *AuthProvider) VerifyTokenAuthority(ctx context.Context, req *auth.VerifyTokenAuthorityRequest) (*auth.VerifyTokenAuthorityResponse, error) {
	if !ap.basicAuthHosts[req.Host] {
		if inner, ok := ap.inner.(auth.AuthServer); ok {
			return inner.VerifyTokenAuthority(ctx, req)
		}
		return nil, status.Errorf(codes.Unavailable, "no token authority for %s", req.Host)
	}

	key := ap.deriveKey(req.Host, req.Salt)
	priv := new([64]byte)
	copy((*priv)[:], key)
	return &auth.VerifyTokenAuthorityResponse{Signed: sign.Sign(nil, req.Payload, priv)}, nil
}

// deriveKey derives a deterministic ed25519 key from host + salt + seed.
func (ap *AuthProvider) deriveKey(host string, salt []byte) ed25519.PrivateKey {
	mac := hmac.New(sha256.New, salt)
	mac.Write(ap.seed)
	mac.Write([]byte(host))
	sum := mac.Sum(nil)
	return ed25519.NewKeyFromSeed(sum[:ed25519.SeedSize])
}

// RegistryHost extracts the host portion from a registry ref
// (e.g. "myorg.registry.cpln.io/gvc/workload:buildcache" → "myorg.registry.cpln.io").
func RegistryHost(ref string) string {
	// Remove tag/digest
	ref = strings.SplitN(ref, "@", 2)[0]
	ref = strings.SplitN(ref, ":", 2)[0]
	// First segment is the host
	parts := strings.SplitN(ref, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
