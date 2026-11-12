package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"oengo-api/internal/apperrors"
)

const (
	ContextKeySub     = "oenexaSub"
	ContextKeyEmail   = "oenexaEmail"
	ContextKeyRoles   = "oenexaRoles"
	ContextKeyKycTier = "oenexaKycTier"
	ContextKeyTokenID = "oenexaTokenId"
	ContextKeyClaims  = "oenexaClaims"
)

// OidcClaims represents verified claims issued by the id.oenexa.com OIDC authority.
type OidcClaims struct {
	Subject   string
	UUID      string
	Email     string
	Issuer    string
	Audience  []string
	Roles     []string
	KycTier   string
	TokenID   string
	TokenType string
	IssuedAt  int64
	ExpiresAt int64
}

type jwkEntry struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwkSetResponse struct {
	Keys []jwkEntry `json:"keys"`
}

// JWKSVerifier validates RS256 OIDC JWTs against an id.oenexa.com JWKS endpoint.
type JWKSVerifier struct {
	jwksURL          string
	expectedIssuer   string
	expectedAudience string
	cacheTTL         time.Duration
	httpClient       *http.Client

	mu        sync.RWMutex
	keys      map[string]*rsa.PublicKey
	lastFetch time.Time
}

// NewJWKSVerifier constructs an RS256 JWKS verifier for the given discovery URL, issuer, and audience.
func NewJWKSVerifier(jwksURL, expectedIssuer, expectedAudience string) *JWKSVerifier {
	return &JWKSVerifier{
		jwksURL:          strings.TrimSpace(jwksURL),
		expectedIssuer:   strings.TrimSpace(expectedIssuer),
		expectedAudience: strings.TrimSpace(expectedAudience),
		cacheTTL:         5 * time.Minute,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
		keys: make(map[string]*rsa.PublicKey),
	}
}

// NewJWKSVerifierFromEnv builds a verifier from OENEXA_JWKS_URL, OENEXA_JWT_ISSUER, and OENEXA_JWT_AUDIENCE.
func NewJWKSVerifierFromEnv() *JWKSVerifier {
	jwksURL := os.Getenv("OENEXA_JWKS_URL")
	issuer := os.Getenv("OENEXA_JWT_ISSUER")
	if strings.TrimSpace(issuer) == "" {
		issuer = "https://id.oenexa.com"
	}
	audience := os.Getenv("OENEXA_JWT_AUDIENCE")
	if strings.TrimSpace(audience) == "" {
		audience = "oenexa-api"
	}
	return NewJWKSVerifier(jwksURL, issuer, audience)
}

// VerifyToken parses, validates signature/issuer/audience/expiration, and returns OidcClaims.
func (v *JWKSVerifier) VerifyToken(tokenString string) (*OidcClaims, error) {
	if strings.TrimSpace(tokenString) == "" {
		return nil, errors.New("empty bearer token")
	}

	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, errors.New("malformed compact JWT")
	}

	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid JWT header encoding: %w", err)
	}
	var header struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid JWT header JSON: %w", err)
	}
	if !strings.EqualFold(header.Alg, "RS256") {
		return nil, fmt.Errorf("unsupported JWT algorithm %q: RS256 required", header.Alg)
	}

	pubKey, err := v.resolvePublicKey(header.Kid)
	if err != nil {
		return nil, err
	}

	sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid JWT signature encoding: %w", err)
	}

	signingInput := parts[0] + "." + parts[1]
	digest := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest[:], sigBytes); err != nil {
		return nil, fmt.Errorf("invalid JWT RS256 signature: %w", err)
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid JWT payload encoding: %w", err)
	}

	claims, err := decodeClaims(payloadBytes)
	if err != nil {
		return nil, err
	}

	if err := v.validateClaims(claims, time.Now().Unix()); err != nil {
		return nil, err
	}
	return claims, nil
}

func (v *JWKSVerifier) validateClaims(claims *OidcClaims, nowUnix int64) error {
	if strings.EqualFold(claims.TokenType, "refresh") {
		return errors.New("refresh token cannot be used as bearer access token")
	}
	if claims.ExpiresAt <= 0 || nowUnix > claims.ExpiresAt {
		return errors.New("token has expired")
	}
	if v.expectedIssuer != "" && claims.Issuer != v.expectedIssuer {
		return fmt.Errorf("invalid token issuer %q", claims.Issuer)
	}
	if v.expectedAudience != "" && len(claims.Audience) > 0 {
		matched := false
		for _, a := range claims.Audience {
			if a == v.expectedAudience {
				matched = true
				break
			}
		}
		if !matched {
			return fmt.Errorf("token audience does not contain %q", v.expectedAudience)
		}
	}
	return nil
}

func (v *JWKSVerifier) resolvePublicKey(kid string) (*rsa.PublicKey, error) {
	if key := v.lookupKey(kid); key != nil && time.Since(v.lastFetchTime()) <= v.cacheTTL {
		return key, nil
	}
	if err := v.refreshJWKS(); err != nil {
		return nil, err
	}
	if key := v.lookupKey(kid); key != nil {
		return key, nil
	}
	return nil, fmt.Errorf("signing key %q not found in JWKS", kid)
}

func (v *JWKSVerifier) lastFetchTime() time.Time {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.lastFetch
}

func (v *JWKSVerifier) lookupKey(kid string) *rsa.PublicKey {
	v.mu.RLock()
	defer v.mu.RUnlock()
	if kid != "" {
		return v.keys[kid]
	}
	for _, k := range v.keys {
		return k
	}
	return nil
}

func (v *JWKSVerifier) refreshJWKS() error {
	if v.jwksURL == "" {
		return errors.New("OENEXA_JWKS_URL is not configured")
	}
	resp, err := v.httpClient.Get(v.jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("JWKS endpoint returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("failed to read JWKS body: %w", err)
	}

	var jwks jwkSetResponse
	if err := json.Unmarshal(body, &jwks); err != nil {
		return fmt.Errorf("invalid JWKS JSON: %w", err)
	}

	parsed := make(map[string]*rsa.PublicKey)
	for _, entry := range jwks.Keys {
		if !strings.EqualFold(entry.Kty, "RSA") || entry.N == "" || entry.E == "" {
			continue
		}
		pub, err := parseRSAPublicKey(entry.N, entry.E)
		if err != nil {
			continue
		}
		keyID := entry.Kid
		if keyID == "" {
			keyID = "default"
		}
		parsed[keyID] = pub
	}
	if len(parsed) == 0 {
		return errors.New("JWKS response contained no valid RSA public keys")
	}

	v.mu.Lock()
	v.keys = parsed
	v.lastFetch = time.Now()
	v.mu.Unlock()
	return nil
}

func parseRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}
	n := new(big.Int).SetBytes(nBytes)
	eInt := new(big.Int).SetBytes(eBytes)
	if !eInt.IsInt64() {
		return nil, errors.New("RSA exponent overflow")
	}
	return &rsa.PublicKey{
		N: n,
		E: int(eInt.Int64()),
	}, nil
}

func decodeClaims(payload []byte) (*OidcClaims, error) {
	var raw struct {
		Sub     string          `json:"sub"`
		UUID    string          `json:"uuid"`
		Email   string          `json:"email"`
		Iss     string          `json:"iss"`
		Aud     json.RawMessage `json:"aud"`
		Roles   []string        `json:"roles"`
		KycTier string          `json:"kyc_tier"`
		Jti     string          `json:"jti"`
		Typ     string          `json:"typ"`
		Iat     int64           `json:"iat"`
		Exp     int64           `json:"exp"`
	}
	if err := json.Unmarshal(payload, &raw); err != nil {
		return nil, fmt.Errorf("invalid JWT payload JSON: %w", err)
	}

	var audience []string
	if len(raw.Aud) > 0 && string(raw.Aud) != "null" {
		var single string
		if err := json.Unmarshal(raw.Aud, &single); err == nil {
			if single != "" {
				audience = []string{single}
			}
		} else {
			_ = json.Unmarshal(raw.Aud, &audience)
		}
	}

	return &OidcClaims{
		Subject:   raw.Sub,
		UUID:      raw.UUID,
		Email:     raw.Email,
		Issuer:    raw.Iss,
		Audience:  audience,
		Roles:     raw.Roles,
		KycTier:   raw.KycTier,
		TokenID:   raw.Jti,
		TokenType: raw.Typ,
		IssuedAt:  raw.Iat,
		ExpiresAt: raw.Exp,
	}, nil
}

// OenexaSSOMiddleware verifies id.oenexa.com Bearer RS256 JWTs and populates Gin identity context keys.
func OenexaSSOMiddleware(verifier *JWKSVerifier, required bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))
		if authHeader == "" {
			if required {
				apperrors.RespondWithError(c, apperrors.NewUnauthorized("missing Authorization Bearer token"))
				c.Abort()
				return
			}
			c.Next()
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			apperrors.RespondWithError(c, apperrors.NewUnauthorized("Authorization header must use Bearer scheme"))
			c.Abort()
			return
		}

		token := strings.TrimSpace(strings.TrimPrefix(authHeader, "Bearer "))
		claims, err := verifier.VerifyToken(token)
		if err != nil {
			apperrors.RespondWithError(c, apperrors.NewUnauthorized("invalid SSO token: "+err.Error()))
			c.Abort()
			return
		}

		sub := claims.UUID
		if sub == "" {
			sub = claims.Subject
		}
		c.Set(ContextKeySub, sub)
		c.Set(ContextKeyEmail, claims.Email)
		c.Set(ContextKeyRoles, claims.Roles)
		c.Set(ContextKeyKycTier, claims.KycTier)
		c.Set(ContextKeyTokenID, claims.TokenID)
		c.Set(ContextKeyClaims, claims)
		c.Next()
	}
}
