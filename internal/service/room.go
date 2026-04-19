package service

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/fistream/fistream/internal/domain"
	"github.com/fistream/fistream/internal/model"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const roomCodeAlphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type RoomStore interface {
	CreateRoom(ctx context.Context, room model.Room) (model.Room, error)
	GetRoomByCode(ctx context.Context, roomCode string) (model.Room, error)
	AddParticipant(ctx context.Context, participant model.Participant) (model.Participant, error)
	TouchRoomExpiry(ctx context.Context, roomID string, expiresAt time.Time) error
	CloseRoom(ctx context.Context, roomCode string, closedAt time.Time) error
	CloseExpiredRooms(ctx context.Context, now time.Time) (int64, error)
}

type Config struct {
	ServiceAccessPassword string
	RoomTTL               time.Duration
	JitsiDomain           string
	JitsiAppID            string
	JitsiAppSecret        string
	JitsiAudience         string
	JitsiSubject          string
	JitsiTokenTTL         time.Duration
	APITokenSecret        string
	APITokenTTL           time.Duration
}

type RoomService struct {
	store  RoomStore
	cfg    Config
	clock  func() time.Time
	random io.Reader
}

type APITokenClaims struct {
	RoomCode      string `json:"room_code"`
	Role          string `json:"role"`
	DisplayName   string `json:"display_name"`
	ParticipantID string `json:"participant_id"`
	jwt.RegisteredClaims
}

func NewRoomService(store RoomStore, cfg Config) *RoomService {
	return &RoomService{store: store, cfg: cfg, clock: time.Now, random: rand.Reader}
}

func (s *RoomService) CreateRoom(ctx context.Context, displayName, servicePassword, clientFingerprint string) (model.RoomAccess, error) {
	if err := s.validateServicePassword(servicePassword); err != nil {
		return model.RoomAccess{}, err
	}
	displayName = normalizeDisplayName(displayName)
	if displayName == "" {
		return model.RoomAccess{}, domain.ErrInvalidDisplayName
	}
	now := s.clock().UTC()
	_, _ = s.store.CloseExpiredRooms(ctx, now)

	expiresAt := now.Add(s.cfg.RoomTTL)
	room := model.Room{
		ID:              uuid.NewString(),
		HostDisplayName: displayName,
		Status:          model.RoomStatusActive,
		CreatedAt:       now,
		ExpiresAt:       expiresAt,
	}

	var created model.Room
	var err error
	for attempt := 0; attempt < 12; attempt++ {
		room.Code, err = s.generateRoomCode(6)
		if err != nil {
			return model.RoomAccess{}, err
		}
		created, err = s.store.CreateRoom(ctx, room)
		if err == nil {
			break
		}
		if err != domain.ErrRoomCodeConflict {
			return model.RoomAccess{}, err
		}
	}
	if err != nil {
		return model.RoomAccess{}, fmt.Errorf("create room after retries: %w", err)
	}

	participant, err := s.store.AddParticipant(ctx, model.Participant{
		ID:                uuid.NewString(),
		RoomID:            created.ID,
		DisplayName:       displayName,
		Role:              "host",
		JoinedAt:          now,
		ClientFingerprint: strings.TrimSpace(clientFingerprint),
	})
	if err != nil {
		return model.RoomAccess{}, err
	}
	return s.issueRoomAccess(created, participant, "host")
}

func (s *RoomService) JoinRoom(ctx context.Context, displayName, roomCode, servicePassword, clientFingerprint string) (model.RoomAccess, error) {
	if err := s.validateServicePassword(servicePassword); err != nil {
		return model.RoomAccess{}, err
	}
	displayName = normalizeDisplayName(displayName)
	if displayName == "" {
		return model.RoomAccess{}, domain.ErrInvalidDisplayName
	}
	roomCode = normalizeRoomCode(roomCode)
	if roomCode == "" {
		return model.RoomAccess{}, domain.ErrInvalidRoomCode
	}

	now := s.clock().UTC()
	room, err := s.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return model.RoomAccess{}, err
	}
	if room.Status == model.RoomStatusClosed {
		return model.RoomAccess{}, domain.ErrRoomClosed
	}
	if !room.ExpiresAt.After(now) {
		_ = s.store.CloseRoom(ctx, roomCode, now)
		return model.RoomAccess{}, domain.ErrRoomExpired
	}

	participant, err := s.store.AddParticipant(ctx, model.Participant{
		ID:                uuid.NewString(),
		RoomID:            room.ID,
		DisplayName:       displayName,
		Role:              "viewer",
		JoinedAt:          now,
		ClientFingerprint: strings.TrimSpace(clientFingerprint),
	})
	if err != nil {
		return model.RoomAccess{}, err
	}

	room.ExpiresAt = now.Add(s.cfg.RoomTTL)
	if err := s.store.TouchRoomExpiry(ctx, room.ID, room.ExpiresAt); err != nil {
		return model.RoomAccess{}, err
	}
	return s.issueRoomAccess(room, participant, "viewer")
}

func (s *RoomService) CloseRoom(ctx context.Context, roomCode, apiToken string) error {
	roomCode = normalizeRoomCode(roomCode)
	if roomCode == "" {
		return domain.ErrInvalidRoomCode
	}
	claims, err := s.ParseAPIToken(apiToken)
	if err != nil {
		return err
	}
	if claims.Role != "host" || normalizeRoomCode(claims.RoomCode) != roomCode {
		return domain.ErrNotHost
	}
	return s.store.CloseRoom(ctx, roomCode, s.clock().UTC())
}

func (s *RoomService) CloseExpiredRooms(ctx context.Context) (int64, error) {
	return s.store.CloseExpiredRooms(ctx, s.clock().UTC())
}

func (s *RoomService) ParseAPIToken(tokenString string) (*APITokenClaims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, domain.ErrInvalidToken
	}
	claims := &APITokenClaims{}
	parsed, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(s.cfg.APITokenSecret), nil
	})
	if err != nil || !parsed.Valid {
		return nil, domain.ErrInvalidToken
	}
	return claims, nil
}

func (s *RoomService) validateServicePassword(given string) error {
	expected := strings.TrimSpace(s.cfg.ServiceAccessPassword)
	if expected == "" {
		return nil
	}
	given = strings.TrimSpace(given)
	if subtle.ConstantTimeCompare([]byte(given), []byte(expected)) != 1 {
		return domain.ErrInvalidServicePassword
	}
	return nil
}

func (s *RoomService) issueRoomAccess(room model.Room, participant model.Participant, role string) (model.RoomAccess, error) {
	jitsiToken, err := s.issueJitsiJWT(room.Code, participant.ID, participant.DisplayName)
	if err != nil {
		return model.RoomAccess{}, err
	}
	apiToken, err := s.issueAPIToken(room.Code, role, participant.DisplayName, participant.ID)
	if err != nil {
		return model.RoomAccess{}, err
	}
	return model.RoomAccess{
		RoomCode:  room.Code,
		ExpiresAt: room.ExpiresAt,
		Role:      role,
		APIToken:  apiToken,
		Jitsi: model.JitsiCredentials{
			Domain:      strings.TrimSpace(s.cfg.JitsiDomain),
			RoomName:    room.Code,
			JWT:         jitsiToken,
			DisplayName: participant.DisplayName,
		},
	}, nil
}

func (s *RoomService) issueAPIToken(roomCode, role, displayName, participantID string) (string, error) {
	now := s.clock().UTC()
	claims := APITokenClaims{
		RoomCode:      roomCode,
		Role:          role,
		DisplayName:   displayName,
		ParticipantID: participantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "fistream-api",
			Audience:  []string{"fistream-api"},
			Subject:   participantID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.cfg.APITokenTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.APITokenSecret))
}

func (s *RoomService) issueJitsiJWT(roomCode, participantID, displayName string) (string, error) {
	now := s.clock().UTC()
	claims := jwt.MapClaims{
		"aud":  s.cfg.JitsiAudience,
		"iss":  s.cfg.JitsiAppID,
		"sub":  s.cfg.JitsiSubject,
		"room": roomCode,
		"exp":  now.Add(s.cfg.JitsiTokenTTL).Unix(),
		"nbf":  now.Unix(),
		"iat":  now.Unix(),
		"context": map[string]any{
			"user": map[string]any{
				"id":   participantID,
				"name": displayName,
			},
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JitsiAppSecret))
}

func (s *RoomService) generateRoomCode(length int) (string, error) {
	if length <= 0 {
		length = 6
	}
	bytes := make([]byte, length)
	if _, err := io.ReadFull(s.random, bytes); err != nil {
		return "", err
	}
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = roomCodeAlphabet[int(bytes[i])%len(roomCodeAlphabet)]
	}
	return string(result), nil
}

func normalizeRoomCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) > 12 {
		value = value[:12]
	}
	for _, char := range value {
		if !((char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9')) {
			return ""
		}
	}
	return value
}

func normalizeDisplayName(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	runes := []rune(value)
	if len(runes) > 40 {
		value = string(runes[:40])
	}
	return value
}

