-- Этот файл подхватит Postgres при первом старте тома
-- (каталог /docker-entrypoint-initdb.d внутри официального образа).
-- Повторные запуски его не выполняют: схема ещё раз создаётся в database.Migrate.

CREATE TABLE IF NOT EXISTS posts (
    id         BIGSERIAL PRIMARY KEY,
    title      TEXT        NOT NULL,
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
