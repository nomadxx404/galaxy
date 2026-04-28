create schema if not exists log;

create table if not exists log.logs (
    log_id bigserial primary key not null,
    created_at timestamptz default now () not null,
    service varchar not null,
    method varchar not null,
    path varchar not null,
    status int not null,
    account_uuid varchar not null,
    request_id varchar not null,
    body_hash varchar not null,
    body_size int not null
);
