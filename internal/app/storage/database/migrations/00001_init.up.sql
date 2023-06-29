CREATE TABLE IF NOT EXISTS url_dependencies
(
    id             serial primary key,
    short_url      varchar(255) not null unique,
    original_url   varchar(255) not null,
    user_id        integer      null unique,
    is_deleted     boolean      default false
);