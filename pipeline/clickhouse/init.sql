-- ClickHouse tables for CDC analytics pipeline
-- This script is mounted into ClickHouse as an init script.

-- Kafka engine table: reads raw Debezium CDC envelope
CREATE TABLE IF NOT EXISTS kafka_encrypted_messages_raw (
    before String,
    after String,
    source String,
    op String
) ENGINE = Kafka
SETTINGS
    kafka_broker_list = 'redpanda:9092',
    kafka_topic_list = 'securecollab.public.encrypted_messages',
    kafka_group_name = 'clickhouse-cdc-consumer',
    kafka_format = 'JSONEachRow',
    kafka_skip_broken_messages = 10;

-- MergeTree destination table for analytics queries
CREATE TABLE IF NOT EXISTS encrypted_messages_analytics (
    id String,
    sender_user_id String,
    recipient_user_id String,
    channel_id String,
    content_type String,
    created_at DateTime64(3),
    ingested_at DateTime64(3) DEFAULT now64(3)
) ENGINE = MergeTree()
ORDER BY (created_at, sender_user_id)
PARTITION BY toYYYYMM(created_at);

-- Materialized view: extracts fields from Debezium "after" envelope
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_encrypted_messages_analytics
TO encrypted_messages_analytics AS
SELECT
    JSONExtractString(after, 'id')                AS id,
    JSONExtractString(after, 'sender_user_id')    AS sender_user_id,
    JSONExtractString(after, 'recipient_user_id') AS recipient_user_id,
    JSONExtractString(after, 'channel_id')        AS channel_id,
    JSONExtractString(after, 'content_type')      AS content_type,
    toDateTime64(JSONExtractInt(after, 'created_at') / 1000, 3) AS created_at
FROM kafka_encrypted_messages_raw
WHERE op IN ('c', 'r', 'u');

-- Hourly aggregation table for the volume endpoint
CREATE TABLE IF NOT EXISTS message_volume_hourly (
    hour DateTime,
    channel_id String,
    sender_user_id String,
    message_count UInt64
) ENGINE = SummingMergeTree(message_count)
ORDER BY (hour, channel_id, sender_user_id)
PARTITION BY toYYYYMM(hour);

CREATE MATERIALIZED VIEW IF NOT EXISTS mv_message_volume_hourly
TO message_volume_hourly AS
SELECT
    toStartOfHour(toDateTime64(JSONExtractInt(after, 'created_at') / 1000, 3)) AS hour,
    JSONExtractString(after, 'channel_id')        AS channel_id,
    JSONExtractString(after, 'sender_user_id')    AS sender_user_id,
    1 AS message_count
FROM kafka_encrypted_messages_raw
WHERE op IN ('c', 'r', 'u');
