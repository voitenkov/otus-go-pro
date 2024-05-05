CREATE TABLE IF NOT EXISTS events
(
    id            uuid PRIMARY KEY,
    user_id       uuid         NOT NULL,
    title         varchar      NOT NULL,
    description   text         NULL,
    start_time    timestamptz  NOT NULL,
    finish_time   timestamptz  NOT NULL,
    notify_before smallint NULL
);