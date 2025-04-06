#!/bin/sh
set -e

echo "Running pre-start configuration..."

if [ -n "$KVDB_DATA_DIR" ]; then
    echo "Configuring data directory: $KVDB_DATA_DIR"
    mkdir -p "$KVDB_DATA_DIR"
    chown kvdb:kvdb "$KVDB_DATA_DIR"
    chmod 750 "$KVDB_DATA_DIR"
fi

if [ -n "$KVDB_LOG_DIR" ]; then
    echo "Configuring log directory: $KVDB_LOG_DIR"
    mkdir -p "$KVDB_LOG_DIR"
    chown kvdb:kvdb "$KVDB_LOG_DIR"
    chmod 755 "$KVDB_LOG_DIR"
fi

echo "Pre-start configuration completed"
exec /sbin/su-exec kvdb "$@"