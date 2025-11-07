#!/bin/bash
set -euo pipefail

REMOTE_HOST="serv-app-04"
REMOTE_DIR="/var/backups/afs-challenge"
RESTORE_DIR="/srv/temp/"

echo "📋 Backups disponibles:"
ssh "$REMOTE_HOST" "ls -1 $REMOTE_DIR/*.tar.gz" | nl

read -p "Número de backup a restaurar: " NUM
BACKUP=$(ssh "$REMOTE_HOST" "ls -1 $REMOTE_DIR/*.tar.gz" | sed -n "${NUM}p")

echo "🔄 Restaurando: $BACKUP"
ssh "$REMOTE_HOST" "cat $BACKUP" | tar xzf - -C "$RESTORE_DIR"

echo "✅ Restauración completa"