-- ClickHouse tables for CDC analytics pipeline
-- This script is mounted into ClickHouse as an init script.

-- Kafka engine table: reads from the Debezium CDC topic
CREATE TABLE IF NOT EXISTS kafka_encrypted_messages (
    id String,
    sender_user_id String,
    recipient_user_id String,
    channel_id String,
    content_type String,
    created_at DateTime64(3)
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

-- Materialized view: auto-moves data from Kafka engine to MergeTree
CREATE MATERIALIZED VIEW IF NOT EXISTS mv_encrypted_messages_analytics
TO encrypted_messages_analytics AS
SELECT
    id,
    sender_user_id,
    recipient_user_id,
    channel_id,
    content_type,
    created_at
FROM kafka_encrypted_messages;

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
    toStartOfHour(created_at) AS hour,
    channel_id,
    sender_user_id,
    1 AS message_count
FROM kafka_encrypted_messages;
