package fasts3

import (
	"encoding/base64"
	"net/http"
	"strings"
)

// AuthKind identifies the authentication strategy applied to outgoing FastS3 HTTP requests.
type AuthKind int

const (
	// AuthNone sends no authentication headers (anonymous mode).
	AuthNone AuthKind = iota
	// AuthBasic sends `Authorization: Basic <base64(user:pass)>`.
	AuthBasic
	// AuthCustomHeaders sends `X-FastS3-User` and `X-FastS3-Pass`.
	AuthCustomHeaders
	// AuthBearer sends `Authorization: Bearer <token>` (Bootstrap Token or API Key).
	AuthBearer
	// AuthAPIKey sends `X-API-Key: <key>`.
	AuthAPIKey
	// AuthFastS3Token sends `X-FastS3-Token: <token>`.
	AuthFastS3Token
)

// Credentials holds authentication details for a FastS3 Client.
type Credentials struct {
	Kind     AuthKind
	Username string
	Password string
	Token    string
}

// AnonymousCredentials returns unauthenticated Credentials.
func AnonymousCredentials() Credentials {
	return Credentials{Kind: AuthNone}
}

// BasicCredentials creates HTTP Basic Authentication credentials.
func BasicCredentials(username, password string) Credentials {
	return Credentials{
		Kind:     AuthBasic,
		Username: username,
		Password: password,
	}
}

// CustomHeaderCredentials creates `X-FastS3-User` / `X-FastS3-Pass` header credentials.
func CustomHeaderCredentials(username, password string) Credentials {
	return Credentials{
		Kind:     AuthCustomHeaders,
		Username: username,
		Password: password,
	}
}

// BearerCredentials creates `Authorization: Bearer <token>` credentials (valid for Bootstrap Tokens and API Keys).
func BearerCredentials(token string) Credentials {
	return Credentials{
		Kind:  AuthBearer,
		Token: strings.TrimSpace(token),
	}
}

// APIKeyCredentials creates `X-API-Key: <apiKey>` credentials.
func APIKeyCredentials(apiKey string) Credentials {
	return Credentials{
		Kind:  AuthAPIKey,
		Token: strings.TrimSpace(apiKey),
	}
}

// FastS3TokenCredentials creates `X-FastS3-Token: <token>` credentials.
func FastS3TokenCredentials(token string) Credentials {
	return Credentials{
		Kind:  AuthFastS3Token,
		Token: strings.TrimSpace(token),
	}
}

// CredentialsFromBasicHeader parses a `"Basic <base64>"` header value returned by login or bootstrap-setup.
func CredentialsFromBasicHeader(headerVal string) (Credentials, bool) {
	trimmed := strings.TrimSpace(headerVal)
	if len(trimmed) >= 6 && strings.EqualFold(trimmed[:6], "basic ") {
		trimmed = strings.TrimSpace(trimmed[6:])
	}
	decoded, err := base64.StdEncoding.DecodeString(trimmed)
	if err != nil {
		return Credentials{}, false
	}
	user, pass, ok := strings.Cut(string(decoded), ":")
	if !ok {
		return Credentials{}, false
	}
	return BasicCredentials(user, pass), true
}

func (c Credentials) apply(req *http.Request) {
	switch c.Kind {
	case AuthNone:
		return
	case AuthBasic:
		encoded := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
		req.Header.Set("Authorization", "Basic "+encoded)
	case AuthCustomHeaders:
		req.Header.Set("X-FastS3-User", c.Username)
		req.Header.Set("X-FastS3-Pass", c.Password)
	case AuthBearer:
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.Token))
	case AuthAPIKey:
		req.Header.Set("X-API-Key", strings.TrimSpace(c.Token))
	case AuthFastS3Token:
		req.Header.Set("X-FastS3-Token", strings.TrimSpace(c.Token))
	}
}

func (c Credentials) toQueryParam() (key, val string, ok bool) {
	switch c.Kind {
	case AuthNone:
		return "", "", false
	case AuthBasic, AuthCustomHeaders:
		encoded := base64.StdEncoding.EncodeToString([]byte(c.Username + ":" + c.Password))
		return "auth", encoded, true
	case AuthBearer, AuthFastS3Token:
		return "token", strings.TrimSpace(c.Token), true
	case AuthAPIKey:
		return "api_key", strings.TrimSpace(c.Token), true
	default:
		return "", "", false
	}
}
