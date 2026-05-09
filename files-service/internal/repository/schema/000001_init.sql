create schema if not exists file;

create table if not exists file.files (
    file_id bigserial primary key not null,
    entity_type varchar not null,
    entity_id varchar,
    name varchar not null,
    size bigint not null,
    content_type varchar not null,
    storage_key varchar not null,
    status varchar default 'PENDING' not null,
    created_by varchar not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null
);

create table if not exists file.outbox_events (
    message_uuid varchar primary key not null,
    account_uuid varchar not null,
    request_id varchar not null,
    event_type varchar not null,
    payload jsonb not null,
    status varchar(20) default 'PENDING' not null,
    retry_count int default 0 not null,
    last_error text,
    created_at timestamptz default now () not null,
    processed_at timestamptz,
    locked_until timestamptz
);
