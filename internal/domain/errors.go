package domain

import "errors"

var (
	ErrInvalidServicePassword = errors.New("invalid_service_password")
	ErrInvalidDisplayName     = errors.New("invalid_display_name")
	ErrInvalidRoomCode        = errors.New("invalid_room_code")
	ErrRoomNotFound           = errors.New("room_not_found")
	ErrRoomClosed             = errors.New("room_closed")
	ErrRoomExpired            = errors.New("room_expired")
	ErrRoomCodeConflict       = errors.New("room_code_conflict")
	ErrNotHost                = errors.New("not_room_host")
	ErrInvalidToken           = errors.New("invalid_token")
)

