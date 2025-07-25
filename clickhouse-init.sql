CREATE TABLE IF NOT EXISTS logs (
    timestamp DateTime,
    level Enum8('info'=1, 'warn'=2, 'error'=3, 'fatal'=4),
    message String,
    service String
) ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)
ORDER BY (service, timestamp); 