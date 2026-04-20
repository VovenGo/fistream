package service

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/fistream/fistream/internal/domain"
	"github.com/fistream/fistream/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

type memoryStore struct {
	mu           sync.Mutex
	roomsByCode  map[string]model.Room
	roomsByID    map[string]model.Room
	participants map[string]model.Participant
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		roomsByCode:  map[string]model.Room{},
		roomsByID:    map[string]model.Room{},
		participants: map[string]model.Participant{},
	}
}

func (m *memoryStore) CreateRoom(_ context.Context, room model.Room) (model.Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.roomsByCode[room.Code]; exists {
		return model.Room{}, domain.ErrRoomCodeConflict
	}
	m.roomsByCode[room.Code] = room
	m.roomsByID[room.ID] = room
	return room, nil
}

func (m *memoryStore) GetRoomByCode(_ context.Context, roomCode string) (model.Room, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	room, ok := m.roomsByCode[strings.ToUpper(roomCode)]
	if !ok {
		return model.Room{}, domain.ErrRoomNotFound
	}
	return room, nil
}

func (m *memoryStore) AddParticipant(_ context.Context, participant model.Participant) (model.Participant, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.participants[participant.ID] = participant
	return participant, nil
}

func (m *memoryStore) TouchRoomExpiry(_ context.Context, roomID string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	room := m.roomsByID[roomID]
	room.ExpiresAt = expiresAt
	m.roomsByID[roomID] = room
	m.roomsByCode[room.Code] = room
	return nil
}

func (m *memoryStore) CloseRoom(_ context.Context, roomCode string, closedAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	room, ok := m.roomsByCode[strings.ToUpper(roomCode)]
	if !ok {
		return domain.ErrRoomNotFound
	}
	room.Status = model.RoomStatusClosed
	room.ClosedAt = &closedAt
	m.roomsByCode[room.Code] = room
	m.roomsByID[room.ID] = room
	return nil
}

func (m *memoryStore) CloseExpiredRooms(_ context.Context, now time.Time) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var closed int64
	for code, room := range m.roomsByCode {
		if room.Status == model.RoomStatusActive && !room.ExpiresAt.After(now) {
			closedAt := now
			room.Status = model.RoomStatusClosed
			room.ClosedAt = &closedAt
			m.roomsByCode[code] = room
			m.roomsByID[room.ID] = room
			closed++
		}
	}
	return closed, nil
}

func defaultServiceForTest(store RoomStore) *RoomService {
	svc := NewRoomService(store, Config{
		ServiceAccessPassword: "secret",
		RoomTTL:               2 * time.Hour,
		JitsiDomain:           "meet.example.com",
		JitsiAppID:            "fistream",
		JitsiAppSecret:        "jitsi-secret",
		JitsiAudience:         "fistream",
		JitsiSubject:          "meet.jitsi",
		JitsiTokenTTL:         2 * time.Hour,
		APITokenSecret:        "api-secret",
		APITokenTTL:           6 * time.Hour,
	})
	baseNow := time.Now().UTC().Add(-time.Minute)
	svc.clock = func() time.Time {
		return baseNow
	}
	return svc
}

func TestCreateRoomRequiresServicePassword(t *testing.T) {
	svc := defaultServiceForTest(newMemoryStore())
	_, err := svc.CreateRoom(context.Background(), "User", "bad-password", "")
	require.ErrorIs(t, err, domain.ErrInvalidServicePassword)
}

func TestCreateAndJoinHappyPath(t *testing.T) {
	svc := defaultServiceForTest(newMemoryStore())

	host, err := svc.CreateRoom(context.Background(), "Host", "secret", "fp-host")
	require.NoError(t, err)
	require.Len(t, host.RoomCode, 6)
	require.Equal(t, "host", host.Role)
	require.NotEmpty(t, host.Jitsi.JWT)
	require.NotEmpty(t, host.APIToken)

	joined, err := svc.JoinRoom(context.Background(), "Guest", host.RoomCode, "secret", "fp-guest")
	require.NoError(t, err)
	require.Equal(t, host.RoomCode, joined.RoomCode)
	require.Equal(t, "viewer", joined.Role)
	require.NotEmpty(t, joined.Jitsi.JWT)
}

func TestJoinExpiredRoomReturnsGone(t *testing.T) {
	store := newMemoryStore()
	svc := defaultServiceForTest(store)

	created, err := svc.CreateRoom(context.Background(), "Host", "secret", "")
	require.NoError(t, err)

	store.mu.Lock()
	room := store.roomsByCode[created.RoomCode]
	room.ExpiresAt = svc.clock().Add(-time.Minute)
	store.roomsByCode[created.RoomCode] = room
	store.roomsByID[room.ID] = room
	store.mu.Unlock()

	_, err = svc.JoinRoom(context.Background(), "Guest", created.RoomCode, "secret", "")
	require.ErrorIs(t, err, domain.ErrRoomExpired)
}

func TestHostCanCloseRoomUsingAPIToken(t *testing.T) {
	svc := defaultServiceForTest(newMemoryStore())
	created, err := svc.CreateRoom(context.Background(), "Host", "secret", "")
	require.NoError(t, err)

	err = svc.CloseRoom(context.Background(), created.RoomCode, created.APIToken)
	require.NoError(t, err)
}

func TestViewerCannotCloseRoom(t *testing.T) {
	svc := defaultServiceForTest(newMemoryStore())
	created, err := svc.CreateRoom(context.Background(), "Host", "secret", "")
	require.NoError(t, err)
	joined, err := svc.JoinRoom(context.Background(), "Guest", created.RoomCode, "secret", "")
	require.NoError(t, err)

	err = svc.CloseRoom(context.Background(), created.RoomCode, joined.APIToken)
	require.ErrorIs(t, err, domain.ErrNotHost)
}

func TestGenerateRoomCodeCharsetAndUniqueness(t *testing.T) {
	svc := defaultServiceForTest(newMemoryStore())
	seen := map[string]struct{}{}

	for i := 0; i < 500; i++ {
		code, err := svc.generateRoomCode(6)
		require.NoError(t, err)
		require.Len(t, code, 6)
		for _, char := range code {
			require.True(t, (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9'))
		}
		if _, exists := seen[code]; exists {
			t.Fatalf("unexpected duplicate room code generated: %s", code)
		}
		seen[code] = struct{}{}
	}
}

func TestJitsiJWTClaims(t *testing.T) {
	svc := defaultServiceForTest(newMemoryStore())
	tokenString, err := svc.issueJitsiJWT("ABC123", "participant-1", "Alice")
	require.NoError(t, err)

	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte("jitsi-secret"), nil
	})
	require.NoError(t, err)
	require.True(t, parsed.Valid)

	claims, ok := parsed.Claims.(jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, "fistream", claims["aud"])
	require.Equal(t, "fistream", claims["iss"])
	require.Equal(t, "meet.jitsi", claims["sub"])
	require.Equal(t, "ABC123", claims["room"])

	contextClaim, ok := claims["context"].(map[string]any)
	require.True(t, ok)
	userClaim, ok := contextClaim["user"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "Alice", userClaim["name"])
}

