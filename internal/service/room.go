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

const (
	roomCodeAlphabet    = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	defaultRoomCodeLen  = 6
	roleHost            = "host"
	roleViewer          = "viewer"
	apiTokenIssuer      = "fistream-api"
	apiTokenAudience    = "fistream-api"
	roomCodeRetryLimit  = 12
	maxNormalizedRoomID = 12
)

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

	normalizedDisplayName, err := validateAndNormalizeDisplayName(displayName)
	if err != nil {
		return model.RoomAccess{}, err
	}

	now := s.nowUTC()
	s.closeExpiredRoomsBestEffort(ctx, now)

	room, err := s.createActiveRoom(ctx, normalizedDisplayName, now)
	if err != nil {
		return model.RoomAccess{}, err
	}

	participant, err := s.addParticipant(ctx, room.ID, normalizedDisplayName, roleHost, now, clientFingerprint)
	if err != nil {
		return model.RoomAccess{}, err
	}

	return s.issueRoomAccess(room, participant, roleHost)
}

func (s *RoomService) JoinRoom(ctx context.Context, displayName, roomCode, servicePassword, clientFingerprint string) (model.RoomAccess, error) {
	if err := s.validateServicePassword(servicePassword); err != nil {
		return model.RoomAccess{}, err
	}

	normalizedDisplayName, err := validateAndNormalizeDisplayName(displayName)
	if err != nil {
		return model.RoomAccess{}, err
	}

	normalizedRoomCode, err := validateAndNormalizeRoomCode(roomCode)
	if err != nil {
		return model.RoomAccess{}, err
	}

	now := s.nowUTC()
	room, err := s.getJoinableRoom(ctx, normalizedRoomCode, now)
	if err != nil {
		return model.RoomAccess{}, err
	}

	participant, err := s.addParticipant(ctx, room.ID, normalizedDisplayName, roleViewer, now, clientFingerprint)
	if err != nil {
		return model.RoomAccess{}, err
	}

	room.ExpiresAt = now.Add(s.cfg.RoomTTL)
	if err := s.store.TouchRoomExpiry(ctx, room.ID, room.ExpiresAt); err != nil {
		return model.RoomAccess{}, err
	}

	return s.issueRoomAccess(room, participant, roleViewer)
}

func (s *RoomService) CloseRoom(ctx context.Context, roomCode, apiToken string) error {
	normalizedRoomCode, err := validateAndNormalizeRoomCode(roomCode)
	if err != nil {
		return err
	}

	claims, err := s.ParseAPIToken(apiToken)
	if err != nil {
		return err
	}
	if claims.Role != roleHost || normalizeRoomCode(claims.RoomCode) != normalizedRoomCode {
		return domain.ErrNotHost
	}

	return s.store.CloseRoom(ctx, normalizedRoomCode, s.nowUTC())
}

func (s *RoomService) CloseExpiredRooms(ctx context.Context) (int64, error) {
	return s.store.CloseExpiredRooms(ctx, s.nowUTC())
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
	if subtle.ConstantTimeCompare([]byte(strings.TrimSpace(given)), []byte(expected)) != 1 {
		return domain.ErrInvalidServicePassword
	}
	return nil
}

func (s *RoomService) createActiveRoom(ctx context.Context, hostDisplayName string, now time.Time) (model.Room, error) {
	baseRoom := model.Room{
		ID:              uuid.NewString(),
		HostDisplayName: hostDisplayName,
		Status:          model.RoomStatusActive,
		CreatedAt:       now,
		ExpiresAt:       now.Add(s.cfg.RoomTTL),
	}

	for attempt := 0; attempt < roomCodeRetryLimit; attempt++ {
		code, err := s.generateRoomCode(defaultRoomCodeLen)
		if err != nil {
			return model.Room{}, err
		}

		candidate := baseRoom
		candidate.Code = code
		created, err := s.store.CreateRoom(ctx, candidate)
		if err == nil {
			return created, nil
		}
		if err != domain.ErrRoomCodeConflict {
			return model.Room{}, err
		}
	}

	return model.Room{}, fmt.Errorf("create room after retries: %w", domain.ErrRoomCodeConflict)
}

func (s *RoomService) getJoinableRoom(ctx context.Context, roomCode string, now time.Time) (model.Room, error) {
	room, err := s.store.GetRoomByCode(ctx, roomCode)
	if err != nil {
		return model.Room{}, err
	}

	if room.Status == model.RoomStatusClosed {
		return model.Room{}, domain.ErrRoomClosed
	}

	if !room.ExpiresAt.After(now) {
		_ = s.store.CloseRoom(ctx, roomCode, now)
		return model.Room{}, domain.ErrRoomExpired
	}

	return room, nil
}

func (s *RoomService) addParticipant(ctx context.Context, roomID, displayName, role string, joinedAt time.Time, clientFingerprint string) (model.Participant, error) {
	participant := model.Participant{
		ID:                uuid.NewString(),
		RoomID:            roomID,
		DisplayName:       displayName,
		Role:              role,
		JoinedAt:          joinedAt,
		ClientFingerprint: strings.TrimSpace(clientFingerprint),
	}
	return s.store.AddParticipant(ctx, participant)
}

func (s *RoomService) closeExpiredRoomsBestEffort(ctx context.Context, now time.Time) {
	_, _ = s.store.CloseExpiredRooms(ctx, now)
}

func (s *RoomService) nowUTC() time.Time {
	return s.clock().UTC()
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
	now := s.nowUTC()
	claims := buildAPITokenClaims(roomCode, role, displayName, participantID, now, s.cfg.APITokenTTL)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.APITokenSecret))
}

func (s *RoomService) issueJitsiJWT(roomCode, participantID, displayName string) (string, error) {
	claims := s.buildJitsiClaims(roomCode, participantID, displayName, s.nowUTC())
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JitsiAppSecret))
}

func (s *RoomService) buildJitsiClaims(roomCode, participantID, displayName string, now time.Time) jwt.MapClaims {
	return jwt.MapClaims{
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
}

func buildAPITokenClaims(roomCode, role, displayName, participantID string, now time.Time, ttl time.Duration) APITokenClaims {
	return APITokenClaims{
		RoomCode:      roomCode,
		Role:          role,
		DisplayName:   displayName,
		ParticipantID: participantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    apiTokenIssuer,
			Audience:  []string{apiTokenAudience},
			Subject:   participantID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
}

func (s *RoomService) generateRoomCode(length int) (string, error) {
	if length <= 0 {
		length = defaultRoomCodeLen
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

func validateAndNormalizeDisplayName(value string) (string, error) {
	normalized := normalizeDisplayName(value)
	if normalized == "" {
		return "", domain.ErrInvalidDisplayName
	}
	return normalized, nil
}

func validateAndNormalizeRoomCode(value string) (string, error) {
	normalized := normalizeRoomCode(value)
	if normalized == "" {
		return "", domain.ErrInvalidRoomCode
	}
	return normalized, nil
}

func normalizeRoomCode(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) > maxNormalizedRoomID {
		value = value[:maxNormalizedRoomID]
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
