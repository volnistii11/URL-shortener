CREATE TABLE IF NOT EXISTS url_dependencies
(
    id
    serial
    primary
    key,
    correlation_id
    varchar
(
    255
) null, short_url varchar
(
    255
) not null unique, original_url varchar
(
    255
) not null unique);