package core_security

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// read private key from REM-file

func LoadEd25519Key(path string) (ed25519.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read key file: %w", err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, errors.New("key file is not valid PEM")
	}

	parsed, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	key, ok := parsed.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not ed25519")
	}
	return key, nil
}

type AccessClaims struct {
	IsAdmin bool `json:"is_admin"`
	jwt.RegisteredClaims
}

type JWTIssuer struct {
	Key    ed25519.PrivateKey
	Kid    string
	Issuer string
	Ttl    time.Duration
}

func NewJWTIssuer(key ed25519.PrivateKey, issuer string, ttl time.Duration) *JWTIssuer {
	pub := key.Public().(ed25519.PublicKey)
	sum := sha256.Sum256(pub)

	return &JWTIssuer{
		Key:    key,
		Kid:    hex.EncodeToString(sum[:8]), // indetificator key for JWKS
		Issuer: issuer,
		Ttl:    ttl,
	}
}

func (i *JWTIssuer) TTL() time.Duration { return i.Ttl }

func (i *JWTIssuer) KeyID() string { return i.Kid }

func (i *JWTIssuer) NewAccessToken(userID uuid.UUID, isAdmin bool) (string, error) {
	now := time.Now()

	claims := AccessClaims{
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    i.Issuer,
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.Ttl)),
			ID:        uuid.NewString(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, claims)
	token.Header["kid"] = i.Kid

	return token.SignedString(i.Key)
}

// return token for client and SHA-256 for database
func NewRefreshToken() (raw string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, err
	}

	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, HashRefreshToken(raw), nil
}

func HashRefreshToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}
