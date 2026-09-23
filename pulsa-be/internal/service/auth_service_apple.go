package service

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"pulsa2/internal/helper"
)

func (s *AuthService) LoginApple(ctx context.Context, email, nama, appleSub, identityToken string) (string, bool, string, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	nama = strings.TrimSpace(nama)
	appleSub = strings.TrimSpace(appleSub)
	identityToken = strings.TrimSpace(identityToken)
	if appleSub == "" || identityToken == "" {
		return "", false, "", errors.New("apple_sub/identity_token required")
	}

	claims, err := verifyAppleIdentityToken(ctx, identityToken)
	if err != nil {
		return "", false, "", err
	}
	if claims.Sub != appleSub {
		return "", false, "", errors.New("apple token subject mismatch")
	}
	if email == "" && claims.Email != "" {
		email = strings.TrimSpace(strings.ToLower(claims.Email))
	}
	if email == "" {
		return "", false, "", errors.New("email required")
	}

	row, err := s.repo.GetByAppleSub(ctx, appleSub)
	if err != nil {
		return "", false, "", err
	}

	created := false
	if row == nil {
		row, err = s.repo.GetByEmail(ctx, email)
		if err != nil {
			return "", false, "", err
		}
		if row != nil {
			if row.AppleSub != "" && row.AppleSub != appleSub {
				return "", false, "", errors.New("email already linked to different apple account")
			}
			if row.AppleSub == "" {
				if err := s.repo.LinkAppleSubByMemberID(ctx, row.ID, appleSub); err != nil {
					return "", false, "", err
				}
				row.AppleSub = appleSub
			}
		}
	}

	if row == nil {
		if nama == "" {
			nama = strings.Split(email, "@")[0]
		}
		row, err = s.repo.CreateAppleMember(ctx, email, nama, appleSub)
		if err != nil {
			return "", false, "", err
		}
		created = true
	}

	if !row.Aktif {
		return "", false, "", errors.New("member inactive")
	}

	role := helper.NormalizeRole(row.Role)
	if role == "" {
		role = helper.RoleUser
	}

	tok, err := helper.MakeJWT(s.jwtSecret, row.ID, role, 7*24*time.Hour)
	if err != nil {
		return "", false, "", err
	}
	return tok, created, role, nil
}

func (s *AuthService) DeactivateAppleMember(ctx context.Context, appleSub string) error {
	return s.repo.DeactivateByAppleSub(ctx, appleSub)
}

type appleTokenClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Iss   string `json:"iss"`
	Exp   int64  `json:"exp"`
}

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleJWKS struct {
	Keys []appleJWK `json:"keys"`
}

var (
	appleKeysCache     *appleJWKS
	appleKeysCacheTime time.Time
	appleKeysMu        sync.RWMutex
)

func verifyAppleIdentityToken(ctx context.Context, token string) (*appleTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid apple identity token format")
	}

	headerJSON, err := base64URLDecode(parts[0])
	if err != nil {
		return nil, errors.New("invalid apple token header")
	}
	var header struct {
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(headerJSON, &header); err != nil {
		return nil, errors.New("invalid apple token header json")
	}

	payloadJSON, err := base64URLDecode(parts[1])
	if err != nil {
		return nil, errors.New("invalid apple token payload")
	}
	var claims appleTokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return nil, errors.New("invalid apple token payload json")
	}

	if claims.Iss != "https://appleid.apple.com" {
		return nil, errors.New("invalid apple token issuer")
	}
	if claims.Exp < time.Now().Unix() {
		return nil, errors.New("apple token expired")
	}

	keys, err := fetchApplePublicKeys(ctx)
	if err != nil {
		return nil, err
	}

	var matchingKey *appleJWK
	for i := range keys.Keys {
		if keys.Keys[i].Kid == header.Kid {
			matchingKey = &keys.Keys[i]
			break
		}
	}
	if matchingKey == nil {
		return nil, errors.New("apple token signing key not found")
	}

	if err := verifyAppleRSASignature(matchingKey, parts[0]+"."+parts[1], parts[2]); err != nil {
		return nil, errors.New("invalid apple token signature")
	}

	return &claims, nil
}

func fetchApplePublicKeys(ctx context.Context) (*appleJWKS, error) {
	appleKeysMu.RLock()
	if appleKeysCache != nil && time.Since(appleKeysCacheTime) < 1*time.Hour {
		defer appleKeysMu.RUnlock()
		return appleKeysCache, nil
	}
	appleKeysMu.RUnlock()

	appleKeysMu.Lock()
	defer appleKeysMu.Unlock()

	if appleKeysCache != nil && time.Since(appleKeysCacheTime) < 1*time.Hour {
		return appleKeysCache, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://appleid.apple.com/auth/keys", nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.New("failed to fetch apple public keys")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("apple public keys endpoint error")
	}

	var jwks appleJWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, errors.New("invalid apple public keys response")
	}

	appleKeysCache = &jwks
	appleKeysCacheTime = time.Now()
	return &jwks, nil
}

func verifyAppleRSASignature(key *appleJWK, message, sig string) error {
	nBytes, err := base64URLDecode(key.N)
	if err != nil {
		return err
	}
	eBytes, err := base64URLDecode(key.E)
	if err != nil {
		return err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	pubKey := &rsa.PublicKey{N: n, E: int(e.Int64())}

	sigBytes, err := base64URLDecode(sig)
	if err != nil {
		return err
	}

	hash := sha256.Sum256([]byte(message))
	return rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, hash[:], sigBytes)
}

func base64URLDecode(s string) ([]byte, error) {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	switch len(s) % 4 {
	case 2:
		s += "=="
	case 3:
		s += "="
	}
	return base64.StdEncoding.DecodeString(s)
}
