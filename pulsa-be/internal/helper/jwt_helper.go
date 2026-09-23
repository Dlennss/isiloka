package helper

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var ErrJWT = errors.New("invalid token")

type JWTClaims struct {
	Sub  int64  `json:"sub"`
	Role string `json:"role"`
	Exp  int64  `json:"exp"`
}

func b64url(b []byte) string {
	return strings.TrimRight(base64.URLEncoding.EncodeToString(b), "=")
}

func signHS256(secret []byte, msg string) string {
	h := hmac.New(sha256.New, secret)
	_, _ = h.Write([]byte(msg))
	return b64url(h.Sum(nil))
}

func MakeJWT(secret []byte, sub int64, role string, ttl time.Duration) (string, error) {
	hdr := b64url([]byte(`{"alg":"HS256","typ":"JWT"}`))
	c := JWTClaims{Sub: sub, Role: role, Exp: time.Now().Add(ttl).Unix()}
	cb, _ := json.Marshal(c)
	payload := b64url(cb)
	msg := hdr + "." + payload
	sig := signHS256(secret, msg)
	return msg + "." + sig, nil
}

func ParseJWT(secret []byte, token string) (JWTClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return JWTClaims{}, ErrJWT
	}
	msg := parts[0] + "." + parts[1]
	if signHS256(secret, msg) != parts[2] {
		return JWTClaims{}, ErrJWT
	}
	raw, err := base64.URLEncoding.WithPadding(base64.NoPadding).DecodeString(parts[1])
	if err != nil {
		return JWTClaims{}, ErrJWT
	}
	var c JWTClaims
	if err := json.Unmarshal(raw, &c); err != nil {
		return JWTClaims{}, ErrJWT
	}
	if time.Now().Unix() > c.Exp {
		return JWTClaims{}, ErrJWT
	}
	return c, nil
}
