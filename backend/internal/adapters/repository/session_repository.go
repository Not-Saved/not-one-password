package repository

import (
	"context"
	"database/sql"
	"fmt"
	"main/internal/core/domain"
	db "main/internal/db/sqlc"
	"net"
	"time"

	"github.com/sqlc-dev/pqtype"
)

type SessionRepository struct {
	queries *db.Queries
}

func NewSessionRepository(dbConn *sql.DB) *SessionRepository {
	return &SessionRepository{
		queries: db.New(dbConn),
	}
}

func (r *SessionRepository) CreateSession(ctx context.Context, userID int32, token string, expiresAt time.Time, userAgent, ipAddress string) (domain.Session, error) {
	dbSession, err := r.queries.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
		UserAgent: sql.NullString{String: userAgent},
		IpAddress: pqtype.Inet{IPNet: net.IPNet{IP: net.IP(ipAddress)}},
	})
	if err != nil {
		return domain.Session{}, err
	}
	return toDomainSession(dbSession), nil
}

func (r *SessionRepository) GetSessionByToken(ctx context.Context, token string) (domain.Session, error) {
	dbSession, err := r.queries.GetSessionByToken(ctx, token)
	if err != nil {
		return domain.Session{}, err
	}
	return toDomainSession(dbSession), nil
}

func (r *SessionRepository) ListActiveSessionsByUser(ctx context.Context, userID int32) ([]domain.Session, error) {
	dbSessions, err := r.queries.ListActiveSessionsByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	sessions := make([]domain.Session, 0, len(dbSessions))
	for _, s := range dbSessions {
		sessions = append(sessions, toDomainSession(s))
	}
	return sessions, nil
}

func (r *SessionRepository) RevokeSessionByToken(ctx context.Context, token string) error {
	return r.queries.RevokeSessionByToken(ctx, token)
}

func (r *SessionRepository) RevokeAllSessionsByUser(ctx context.Context, userID int32) error {
	return r.queries.RevokeAllSessionsByUser(ctx, userID)
}

func toDomainSession(s db.Session) domain.Session {
	return domain.Session{
		ID:        s.ID.String(),
		UserID:    s.UserID,
		Token:     s.Token,
		CreatedAt: s.CreatedAt,
		ExpiresAt: s.ExpiresAt,
		RevokedAt: s.RevokedAt.Time,
		UserAgent: s.UserAgent.String,
		IpAddress: fmt.Sprint(s.IpAddress),
	}
}
