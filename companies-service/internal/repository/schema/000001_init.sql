create schema if not exists company;

create table if not exists company.companies (
    company_uuid varchar primary key not null,
    name varchar not null,
    description varchar,
    logo_file_id int,
    billing_status varchar default 'free' not null,
    billing_until timestamptz,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null
);

create table if not exists company.plans (
    plan_id serial primary key not null,
    name varchar not null,
    description varchar not null,
    price numeric(10, 2) not null,
    period int not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null
);

create table if not exists company.roles (
    role_id serial primary key not null,
    name varchar not null,
    description varchar,
    color VARCHAR(7) not null,
    company_uuid varchar not null references company.companies (company_uuid),
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null
);

create table if not exists company.members (
    company_uuid varchar not null references company.companies (company_uuid),
    account_uuid varchar not null,
    role_id int not null references company.roles (role_id),
    is_owner boolean default false not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    primary key (company_uuid, account_uuid)
);

create table if not exists company.permissions (
    company_uuid varchar not null references company.companies (company_uuid),
    account_uuid varchar not null,
    domain varchar not null,
    mask bigint not null,
    change_by varchar not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    primary key (company_uuid, account_uuid, domain)
);

create table if not exists company.outbox_events (
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
