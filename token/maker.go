package token

import (
	"time"

	"github.com/google/uuid"
)

type Maker interface {
	CreateToken(username string, duration time.Duration) (string, uuid.UUID, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
