# Hey, url-shortener is my first golang project =)

Here I was trained with some new things for the first time.

## Frameworks, libs and tools

- Gin;
- Pgx;
- Golang-jwt;
- Zap;
- Golang-migrate;
- Squirrel;
- Sqlx;
- Pgx.

## Storage options

- Memory;
- File .json;
- PostgreSQL.

## REST API Handlers

| Method | Path                                                  | Description              |
| :----: | :---------------------------------------------------: | :----------------------: |
| POST   | [/]                                            | Create new short url         |
| GET    | [/{short_url}]                                   | Get original url |
| GET    | [/ping]                                   | Ping database server |
| POST    | [/api/shorten]                                   | Crate new short url |
| POST    | [/api/shorten/batch]                                   | Crate batch url |
| GET    | [/api/user/urls]                                   | Get all user urls|
| DELETE    | [/api/user/urls]                                   | Delete batch url|
