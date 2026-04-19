create table if not exists rooms (
    id uuid primary key,
    code varchar(12) not null unique,
    host_display_name varchar(40) not null,
    status varchar(20) not null check (status in ('active', 'closed')),
    created_at timestamptz not null,
    expires_at timestamptz not null,
    closed_at timestamptz
);

create table if not exists room_participants (
    id uuid primary key,
    room_id uuid not null references rooms(id) on delete cascade,
    display_name varchar(40) not null,
    role varchar(20) not null check (role in ('host', 'viewer')),
    joined_at timestamptz not null,
    left_at timestamptz,
    client_fingerprint varchar(255)
);

create index if not exists idx_rooms_code on rooms(code);
create index if not exists idx_rooms_status_expires_at on rooms(status, expires_at);
create index if not exists idx_room_participants_room_id on room_participants(room_id);

