create schema if not exists auth;

create schema if not exists company;

create schema if not exists project;

create schema if not exists task;

create schema if not exists comment;

create schema if not exists file;

create schema if not exists notification;

--------------------------------------auth-----------------------------------------------
create table if not exists auth.account (
    account_uuid varchar primary key not null,
    name varchar not null,
    nickname varchar not null,
    email varchar not null,
    password_hash varchar not null,
    avatar_file_id int,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_verified boolean default false not null,
    is_active boolean default true not null
);

create table if not exists auth.outbox_events (
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

create index if not exists idx_outbox_events_unprocessed on auth.outbox_events (created_at)
where
    status = 'PENDING';

--------------------------------------companies-----------------------------------------------
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

create index if not exists idx_outbox_events_unprocessed on company.outbox_events (created_at)
where
    status = 'PENDING';

--------------------------------------projects-----------------------------------------------
create table if not exists project.projects (
    project_uuid varchar primary key not null,
    company_uuid varchar not null,
    name varchar not null,
    project_key varchar(10) not null,
    description varchar,
    date_start timestamptz,
    date_end timestamptz,
    logo_file_id int,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    unique (company_uuid, project_key)
);

create table if not exists project.roles (
    role_id serial primary key not null,
    name varchar not null,
    description varchar,
    color VARCHAR(7) not null,
    project_uuid varchar not null references project.projects (project_uuid),
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    unique (project_uuid, name)
);

create table if not exists project.members (
    project_uuid varchar not null references project.projects (project_uuid),
    account_uuid varchar not null,
    role_id int not null references project.roles (role_id),
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    primary key (project_uuid, account_uuid)
);

create table if not exists project.permissions (
    project_uuid varchar not null references project.projects (project_uuid),
    account_uuid varchar not null,
    domain varchar not null,
    mask bigint not null,
    change_by varchar not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    primary key (project_uuid, account_uuid, domain)
);

--------------------------------------tasks-----------------------------------------------
create table if not exists task.statuses (
    status_id serial primary key not null,
    project_uuid varchar not null,
    name varchar not null,
    position int not null,
    description varchar,
    color VARCHAR(7) not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    unique (project_uuid, position),
    unique (project_uuid, name)
);

create table if not exists task.tags (
    tag_id serial primary key not null,
    project_uuid varchar not null,
    name varchar not null,
    description varchar,
    color VARCHAR(7) not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    unique (project_uuid, name)
);

create table if not exists task.priorities (
    priority_id serial primary key not null,
    project_uuid varchar not null,
    name varchar not null,
    description varchar,
    color VARCHAR(7) not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    unique (project_uuid, name)
);

create table if not exists task.tasks (
    task_uuid varchar primary key not null,
    project_uuid varchar not null,
    number int not null,
    name varchar not null,
    description varchar,
    status_id int not null references task.statuses (status_id),
    priority_id int not null references task.priorities (priority_id),
    planned_start timestamptz,
    due_date timestamptz,
    completed_at timestamptz,
    created_by varchar not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    unique (project_uuid, number)
);

create table if not exists task.assignments (
    task_uuid varchar not null references task.tasks (task_uuid),
    account_uuid varchar not null,
    is_executor boolean default false,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    primary key (task_uuid, account_uuid)
);

create table if not exists task.task_approval (
    task_uuid varchar primary key not null,
    approved boolean,
    approved_by varchar,
    approved_at timestamptz,
    rejected_reason text,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null
);

create table if not exists task.permissions (
    project_uuid varchar not null,
    account_uuid varchar not null,
    domain varchar not null,
    mask bigint not null,
    change_by varchar not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    primary key (project_uuid, account_uuid, domain)
);

--------------------------------------comments-----------------------------------------------
create table if not exists comment.comments (
    comment_uuid varchar primary key not null,
    task_uuid varchar not null,
    text text not null,
    created_by varchar not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null
);

create table if not exists comment.comments_files (
    comment_uuid varchar not null references comment.comments (comment_uuid),
    file_id int not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    primary key (comment_uuid, file_id)
);

create table if not exists comment.permissions (
    project_uuid varchar not null,
    account_uuid varchar not null,
    domain varchar not null,
    mask bigint not null,
    created_by varchar not null,
    created_at timestamptz default now () not null,
    updated_at timestamptz default now () not null,
    is_active boolean default true not null,
    primary key (project_uuid, account_uuid, domain)
);

--------------------------------------files-----------------------------------------------
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
    account_uuid varchar,
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

create index if not exists idx_outbox_events_unprocessed on file.outbox_events (created_at)
where
    status = 'PENDING';

--------------------------------------notification-----------------------------------------------
create table if not exists notification.notifications (
    notification_id bigserial primary key not null,
    account_uuid varchar not null,
    type varchar not null,
    title varchar not null,
    message text not null,
    entity_type varchar,
    entity_id varchar,
    metadata jsonb,
    is_read boolean default false not null,
    read_at timestamptz,
    created_at timestamptz default now () not null
);
