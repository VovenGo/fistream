package model

import "time"

type RoomStatus string

const (
	RoomStatusActive RoomStatus = "active"
	RoomStatusClosed RoomStatus = "closed"
)

type Room struct {
	ID              string
	Code            string
	HostDisplayName string
	Status          RoomStatus
	CreatedAt       time.Time
	ExpiresAt       time.Time
	ClosedAt        *time.Time
}

type Participant struct {
	ID                string
	RoomID            string
	DisplayName       string
	Role              string
	JoinedAt          time.Time
	LeftAt            *time.Time
	ClientFingerprint string
}

type JitsiCredentials struct {
	Domain      string `json:"domain"`
	RoomName    string `json:"room_name"`
	JWT         string `json:"jwt"`
	DisplayName string `json:"display_name"`
}

type RoomAccess struct {
	RoomCode  string           `json:"room_code"`
	ExpiresAt time.Time        `json:"expires_at"`
	Role      string           `json:"role"`
	APIToken  string           `json:"api_token"`
	Jitsi     JitsiCredentials `json:"jitsi"`
}

