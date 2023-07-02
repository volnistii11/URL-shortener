BEGIN;
CREATE TABLE IF NOT EXISTS url_dependencies
(
    id           serial primary key,
    short_url    varchar(255) not null unique,
    original_url varchar(255) not null,
    user_id      integer      null,
    is_deleted   boolean default false
);

CREATE INDEX url_dependencies_user_id_idx on url_dependencies (user_id);
CREATE INDEX url_dependencies_is_deleted_idx on url_dependencies(is_deleted);
COMMIT;