package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/fistream/fistream/internal/domain"
	"github.com/fistream/fistream/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateRoom(ctx context.Context, room model.Room) (model.Room, error) {
	const query = `
		insert into rooms (id, code, host_display_name, status, created_at, expires_at)
		values ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.pool.Exec(ctx, query, room.ID, room.Code, room.HostDisplayName, room.Status, room.CreatedAt, room.ExpiresAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Room{}, domain.ErrRoomCodeConflict
		}
		return model.Room{}, err
	}
	return room, nil
}

func (s *Store) GetRoomByCode(ctx context.Context, roomCode string) (model.Room, error) {
	const query = `
		select id, code, host_display_name, status, created_at, expires_at, closed_at
		from rooms
		where code = $1
		limit 1
	`
	var room model.Room
	var status string
	err := s.pool.QueryRow(ctx, query, strings.ToUpper(roomCode)).Scan(
		&room.ID,
		&room.Code,
		&room.HostDisplayName,
		&status,
		&room.CreatedAt,
		&room.ExpiresAt,
		&room.ClosedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.Room{}, domain.ErrRoomNotFound
	}
	if err != nil {
		return model.Room{}, err
	}
	room.Status = model.RoomStatus(status)
	return room, nil
}

func (s *Store) AddParticipant(ctx context.Context, participant model.Participant) (model.Participant, error) {
	const query = `
		insert into room_participants (id, room_id, display_name, role, joined_at, client_fingerprint)
		values ($1, $2, $3, $4, $5, $6)
	`
	_, err := s.pool.Exec(ctx, query,
		participant.ID,
		participant.RoomID,
		participant.DisplayName,
		participant.Role,
		participant.JoinedAt,
		participant.ClientFingerprint,
	)
	if err != nil {
		return model.Participant{}, err
	}
	return participant, nil
}

func (s *Store) TouchRoomExpiry(ctx context.Context, roomID string, expiresAt time.Time) error {
	const query = `
		update rooms
		set expires_at = $2
		where id = $1 and status = 'active'
	`
	_, err := s.pool.Exec(ctx, query, roomID, expiresAt)
	return err
}

func (s *Store) CloseRoom(ctx context.Context, roomCode string, closedAt time.Time) error {
	const query = `
		update rooms
		set status = 'closed', closed_at = $2
		where code = $1 and status <> 'closed'
	`
	result, err := s.pool.Exec(ctx, query, strings.ToUpper(roomCode), closedAt)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return domain.ErrRoomNotFound
	}
	return nil
}

func (s *Store) CloseExpiredRooms(ctx context.Context, now time.Time) (int64, error) {
	const query = `
		update rooms
		set status = 'closed', closed_at = $1
		where status = 'active' and expires_at <= $1
	`
	result, err := s.pool.Exec(ctx, query, now)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected(), nil
}

