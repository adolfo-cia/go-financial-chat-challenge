package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/google/uuid"
	"github.com/o1egl/paseto"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issuedAt"`
	ExpiredAt time.Time `json:"expiredAt"`
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	payload := &Payload{
		ID:        tokenId,
		Username:  username,
		IssuedAt:  now,
		ExpiredAt: now.Add(duration),
	}

	return payload, nil
}

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

// NewPasetoMaker creates a new PasetoMaker
func NewPasetoMaker(symmetricKey string) (*PasetoMaker, error) {
	if len(symmetricKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalida key size: must be exact %d characters", chacha20poly1305.KeySize)
	}

	pMaker := &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}

	return pMaker, nil
}

// CreateToken create a new token for a specific username and duration
func (pMaker *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	tokenId, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	iat := time.Now()
	exp := iat.Add(duration)
	jsonToken := paseto.JSONToken{
		Subject:    tokenId.String(),
		IssuedAt:   iat,
		Expiration: exp,
	}
	jsonToken.Set("username", username)

	return pMaker.paseto.Encrypt(pMaker.symmetricKey, jsonToken, nil)
}

// VerifyToken checks if the token is valid or not
func (pMaker *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	jsonToken := &paseto.JSONToken{}
	err := pMaker.paseto.Decrypt(token, pMaker.symmetricKey, jsonToken, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = jsonToken.Validate()
	if err != nil {
		if strings.ContainsAny(err.Error(), "expired") {
			return nil, ErrExpiredToken
		}
		return nil, err
	}

	id, err := uuid.Parse(jsonToken.Subject)
	if err != nil {
		return nil, err
	}
	return &Payload{
		ID:        id,
		Username:  jsonToken.Get("username"),
		IssuedAt:  jsonToken.IssuedAt,
		ExpiredAt: jsonToken.Expiration,
	}, nil
}
