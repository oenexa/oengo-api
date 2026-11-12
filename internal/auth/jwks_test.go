package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func signTestRS256JWT(t *testing.T, priv *rsa.PrivateKey, header map[string]any, payload map[string]any) string {
	t.Helper()
	hBytes, err := json.Marshal(header)
	if err != nil {
		t.Fatalf("marshal header: %v", err)
	}
	pBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	hEnc := base64.RawURLEncoding.EncodeToString(hBytes)
	pEnc := base64.RawURLEncoding.EncodeToString(pBytes)
	signingInput := hEnc + "." + pEnc
	digest := sha256.Sum256([]byte(signingInput))

	sig, err := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatalf("sign RS256: %v", err)
	}
	return signingInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}

func startMockJWKSServer(t *testing.T, pub *rsa.PublicKey, kid string) *httptest.Server {
	t.Helper()
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	jwksPayload := map[string]any{
		"keys": []map[string]any{
			{
				"kty": "RSA",
				"use": "sig",
				"alg": "RS256",
				"kid": kid,
				"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
				"e":   base64.RawURLEncoding.EncodeToString(eBytes),
			},
		},
	}

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/.well-known/jwks.json" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwksPayload)
	}))
}

func TestJWKSVerifier_ValidTokenAndClaims(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	srv := startMockJWKSServer(t, &priv.PublicKey, "oenexa-id-k1")
	defer srv.Close()

	verifier := NewJWKSVerifier(srv.URL+"/.well-known/jwks.json", "https://id.oenexa.com", "oenexa-api")
	now := time.Now().Unix()

	token := signTestRS256JWT(t, priv,
		map[string]any{"alg": "RS256", "kid": "oenexa-id-k1", "typ": "JWT"},
		map[string]any{
			"sub":      "rider@oenexa.com",
			"uuid":     "uuid-rider-777",
			"email":    "rider@oenexa.com",
			"iss":      "https://id.oenexa.com",
			"aud":      []string{"oenexa-api"},
			"roles":    []string{"ROLE_USER", "ROLE_DRIVER"},
			"kyc_tier": "TIER2",
			"jti":      "jti-abc-123",
			"iat":      now,
			"exp":      now + 900,
		},
	)

	claims, err := verifier.VerifyToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.UUID != "uuid-rider-777" || claims.KycTier != "TIER2" || len(claims.Roles) != 2 {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWKSVerifier_RejectsExpiredWrongIssuerBadAlgAndRefresh(t *testing.T) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}
	srv := startMockJWKSServer(t, &priv.PublicKey, "oenexa-id-k1")
	defer srv.Close()

	verifier := NewJWKSVerifier(srv.URL+"/.well-known/jwks.json", "https://id.oenexa.com", "oenexa-api")
	now := time.Now().Unix()

	// 1. Expired token
	expired := signTestRS256JWT(t, priv,
		map[string]any{"alg": "RS256", "kid": "oenexa-id-k1"},
		map[string]any{"sub": "u1", "iss": "https://id.oenexa.com", "exp": now - 10},
	)
	if _, err := verifier.VerifyToken(expired); err == nil {
		t.Fatal("expected error for expired token")
	}

	// 2. Wrong issuer
	wrongIss := signTestRS256JWT(t, priv,
		map[string]any{"alg": "RS256", "kid": "oenexa-id-k1"},
		map[string]any{"sub": "u1", "iss": "https://evil.example.com", "exp": now + 300},
	)
	if _, err := verifier.VerifyToken(wrongIss); err == nil {
		t.Fatal("expected error for wrong issuer")
	}

	// 3. Refresh token rejected as access bearer token
	refreshTok := signTestRS256JWT(t, priv,
		map[string]any{"alg": "RS256", "kid": "oenexa-id-k1"},
		map[string]any{"sub": "u1", "iss": "https://id.oenexa.com", "typ": "refresh", "exp": now + 300},
	)
	if _, err := verifier.VerifyToken(refreshTok); err == nil {
		t.Fatal("expected error when refresh token used as access token")
	}

	// 4. Algorithm confusion (HS256 / none rejected)
	hsHeader := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","kid":"oenexa-id-k1"}`))
	hsPayload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"u1"}`))
	if _, err := verifier.VerifyToken(hsHeader + "." + hsPayload + ".fakesig"); err == nil {
		t.Fatal("expected error for non-RS256 algorithm")
	}

	// 5. Wrong signing key
	otherPriv, _ := rsa.GenerateKey(rand.Reader, 2048)
	forged := signTestRS256JWT(t, otherPriv,
		map[string]any{"alg": "RS256", "kid": "oenexa-id-k1"},
		map[string]any{"sub": "u1", "iss": "https://id.oenexa.com", "exp": now + 300},
	)
	if _, err := verifier.VerifyToken(forged); err == nil {
		t.Fatal("expected error for forged signature")
	}
}

func TestOenexaSSOMiddleware_PopulatesContextAndRejectsInvalid(t *testing.T) {
	gin.SetMode(gin.TestMode)
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	srv := startMockJWKSServer(t, &priv.PublicKey, "oenexa-id-k1")
	defer srv.Close()

	verifier := NewJWKSVerifier(srv.URL+"/.well-known/jwks.json", "https://id.oenexa.com", "oenexa-api")
	now := time.Now().Unix()

	validTok := signTestRS256JWT(t, priv,
		map[string]any{"alg": "RS256", "kid": "oenexa-id-k1"},
		map[string]any{
			"sub":      "user@oenexa.com",
			"uuid":     "sub-999",
			"email":    "user@oenexa.com",
			"iss":      "https://id.oenexa.com",
			"aud":      "oenexa-api",
			"roles":    []string{"ROLE_TRADER"},
			"kyc_tier": "TIER3",
			"jti":      "tok-xyz",
			"exp":      now + 600,
		},
	)

	r := gin.New()
	r.Use(OenexaSSOMiddleware(verifier, true))
	r.GET("/protected", func(c *gin.Context) {
		sub, _ := c.Get(ContextKeySub)
		roles, _ := c.Get(ContextKeyRoles)
		kycTier, _ := c.Get(ContextKeyKycTier)
		c.JSON(http.StatusOK, gin.H{
			"sub":     sub,
			"roles":   roles,
			"kycTier": kycTier,
		})
	})

	// Missing header -> 401
	wMissing := httptest.NewRecorder()
	reqMissing, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(wMissing, reqMissing)
	if wMissing.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when Authorization missing, got %d", wMissing.Code)
	}

	// Valid Bearer token -> 200 with populated context
	wOK := httptest.NewRecorder()
	reqOK, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	reqOK.Header.Set("Authorization", "Bearer "+validTok)
	r.ServeHTTP(wOK, reqOK)
	if wOK.Code != http.StatusOK {
		t.Fatalf("expected 200 with valid SSO JWT, got %d body=%s", wOK.Code, wOK.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(wOK.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp["sub"] != "sub-999" || resp["kycTier"] != "TIER3" {
		t.Fatalf("unexpected context payload: %+v", resp)
	}
}
