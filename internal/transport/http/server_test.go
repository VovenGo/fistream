package httptransport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fistream/fistream/internal/config"
	"github.com/fistream/fistream/internal/domain"
	"github.com/fistream/fistream/internal/model"
	"github.com/fistream/fistream/internal/service"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	store := newServiceTestStore()
	svc := service.NewRoomService(store, service.Config{
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

	srv := NewServer(config.Config{AllowedOrigins: []string{"*"}, JitsiDomain: "meet.example.com", AppBuildID: "test"}, svc)
	return httptest.NewServer(srv.Routes())
}

func TestCreateRoomInvalidPassword(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	body := []byte(`{"display_name":"Host","service_password":"wrong"}`)
	resp, err := http.Post(ts.URL+"/api/v1/rooms/create", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestCreateJoinAndCloseFlow(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	createPayload := []byte(`{"display_name":"Host","service_password":"secret"}`)
	createResp, err := http.Post(ts.URL+"/api/v1/rooms/create", "application/json", bytes.NewReader(createPayload))
	require.NoError(t, err)
	defer createResp.Body.Close()
	require.Equal(t, http.StatusCreated, createResp.StatusCode)

	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	roomCode := created["room_code"].(string)
	hostToken := created["api_token"].(string)

	joinBody := []byte(`{"display_name":"Guest","room_code":"` + roomCode + `","service_password":"secret"}`)
	joinResp, err := http.Post(ts.URL+"/api/v1/rooms/join", "application/json", bytes.NewReader(joinBody))
	require.NoError(t, err)
	defer joinResp.Body.Close()
	require.Equal(t, http.StatusOK, joinResp.StatusCode)

	closeReq, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/rooms/"+roomCode+"/close", bytes.NewReader(nil))
	require.NoError(t, err)
	closeReq.Header.Set("Authorization", "Bearer "+hostToken)
	closeResp, err := http.DefaultClient.Do(closeReq)
	require.NoError(t, err)
	defer closeResp.Body.Close()
	require.Equal(t, http.StatusOK, closeResp.StatusCode)

	joinAfterClose := []byte(`{"display_name":"Other","room_code":"` + roomCode + `","service_password":"secret"}`)
	joinAfterCloseResp, err := http.Post(ts.URL+"/api/v1/rooms/join", "application/json", bytes.NewReader(joinAfterClose))
	require.NoError(t, err)
	defer joinAfterCloseResp.Body.Close()
	require.Equal(t, http.StatusGone, joinAfterCloseResp.StatusCode)
}

func TestCloseRoomForbiddenForViewer(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.Close()

	createPayload := []byte(`{"display_name":"Host","service_password":"secret"}`)
	createResp, err := http.Post(ts.URL+"/api/v1/rooms/create", "application/json", bytes.NewReader(createPayload))
	require.NoError(t, err)
	defer createResp.Body.Close()

	var created map[string]any
	require.NoError(t, json.NewDecoder(createResp.Body).Decode(&created))
	roomCode := created["room_code"].(string)

	joinBody := []byte(`{"display_name":"Guest","room_code":"` + roomCode + `","service_password":"secret"}`)
	joinResp, err := http.Post(ts.URL+"/api/v1/rooms/join", "application/json", bytes.NewReader(joinBody))
	require.NoError(t, err)
	defer joinResp.Body.Close()

	var joined map[string]any
	require.NoError(t, json.NewDecoder(joinResp.Body).Decode(&joined))
	viewerToken := joined["api_token"].(string)

	closeReq, err := http.NewRequest(http.MethodPost, ts.URL+"/api/v1/rooms/"+roomCode+"/close", bytes.NewReader(nil))
	require.NoError(t, err)
	closeReq.Header.Set("Authorization", "Bearer "+viewerToken)
	closeResp, err := http.DefaultClient.Do(closeReq)
	require.NoError(t, err)
	defer closeResp.Body.Close()
	require.Equal(t, http.StatusForbidden, closeResp.StatusCode)
}

// local clone of in-memory store for transport tests
// to avoid circular package imports with service tests.
type serviceTestStore struct {
	roomsByCode  map[string]serviceTestRoom
	roomsByID    map[string]serviceTestRoom
	participants map[string]struct{}
}

type serviceTestRoom struct {
	ID              string
	Code            string
	HostDisplayName string
	Status          string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	ClosedAt        *time.Time
}

func newServiceTestStore() *serviceTestStore {
	return &serviceTestStore{
		roomsByCode:  map[string]serviceTestRoom{},
		roomsByID:    map[string]serviceTestRoom{},
		participants: map[string]struct{}{},
	}
}

func (m *serviceTestStore) CreateRoom(_ context.Context, room model.Room) (model.Room, error) {
	if _, exists := m.roomsByCode[room.Code]; exists {
		return model.Room{}, domain.ErrRoomCodeConflict
	}
	m.roomsByCode[room.Code] = serviceTestRoom{ID: room.ID, Code: room.Code, HostDisplayName: room.HostDisplayName, Status: string(room.Status), CreatedAt: room.CreatedAt, ExpiresAt: room.ExpiresAt}
	m.roomsByID[room.ID] = m.roomsByCode[room.Code]
	return room, nil
}

func (m *serviceTestStore) GetRoomByCode(_ context.Context, roomCode string) (model.Room, error) {
	room, ok := m.roomsByCode[strings.ToUpper(roomCode)]
	if !ok {
		return model.Room{}, domain.ErrRoomNotFound
	}
	return model.Room{ID: room.ID, Code: room.Code, HostDisplayName: room.HostDisplayName, Status: model.RoomStatus(room.Status), CreatedAt: room.CreatedAt, ExpiresAt: room.ExpiresAt, ClosedAt: room.ClosedAt}, nil
}

func (m *serviceTestStore) AddParticipant(_ context.Context, participant model.Participant) (model.Participant, error) {
	m.participants[participant.ID] = struct{}{}
	return participant, nil
}

func (m *serviceTestStore) TouchRoomExpiry(_ context.Context, roomID string, expiresAt time.Time) error {
	room := m.roomsByID[roomID]
	room.ExpiresAt = expiresAt
	m.roomsByID[roomID] = room
	m.roomsByCode[room.Code] = room
	return nil
}

func (m *serviceTestStore) CloseRoom(_ context.Context, roomCode string, closedAt time.Time) error {
	room, ok := m.roomsByCode[strings.ToUpper(roomCode)]
	if !ok {
		return domain.ErrRoomNotFound
	}
	room.Status = "closed"
	room.ClosedAt = &closedAt
	m.roomsByCode[room.Code] = room
	m.roomsByID[room.ID] = room
	return nil
}

func (m *serviceTestStore) CloseExpiredRooms(_ context.Context, now time.Time) (int64, error) {
	var closed int64
	for code, room := range m.roomsByCode {
		if room.Status == "active" && !room.ExpiresAt.After(now) {
			room.Status = "closed"
			closeTS := now
			room.ClosedAt = &closeTS
			m.roomsByCode[code] = room
			m.roomsByID[room.ID] = room
			closed++
		}
	}
	return closed, nil
}

