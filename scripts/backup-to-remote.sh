#!/bin/bash
set -euo pipefail

# # Hacer backup
# chmod +x scripts/backup-to-remote.sh
# ./scripts/backup-to-remote.sh

# # Restaurar
# chmod +x scripts/restore-from-remote.sh
# ./scripts/restore-from-remote.sh

# # Automatizar (cron diario a las 2 AM)
# crontab -e
# 0 2 * * * /srv/afs-challenge/scripts/backup-to-remote.sh >> /var/log/afs-backup.log 2>&1

# ============================================
# Configuración
# ============================================
SOURCE_DIR="/srv/afs-challenge"
REMOTE_HOST="serv-app-04"  # Del ~/.ssh/config
REMOTE_DIR="/var/backups/afs-challenge"
BACKUP_NAME="afs-backup-$(date +%Y%m%d-%H%M%S).tar.gz"
KEEP_DAYS=7  # Retener backups de últimos 7 días

# ============================================
# Colores
# ============================================
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}🔄 AFS Challenge - Remote Backup${NC}"
echo "===================================="

# ============================================
# Verificar conexión SSH
# ============================================
echo -e "${YELLOW}📡 Verificando conexión a $REMOTE_HOST...${NC}"
if ! ssh -o ConnectTimeout=5 "$REMOTE_HOST" "exit" 2>/dev/null; then
    echo -e "${RED}❌ No se puede conectar a $REMOTE_HOST${NC}"
    exit 1
fi
echo -e "${GREEN}✅ Conexión exitosa${NC}"

# ============================================
# Crear archivo de exclusiones temporal
# ============================================
EXCLUDE_FILE=$(mktemp)
cat > "$EXCLUDE_FILE" << 'EOF'
node_modules/
.git/
tmp/
*.log
.env.local
.DS_Store
dist/
build/
coverage/
internal.backup.*
EOF

# ============================================
# Comprimir y transferir
# ============================================
echo -e "${YELLOW}📦 Comprimiendo proyecto...${NC}"

tar czf - \
    -C "$(dirname $SOURCE_DIR)" \
    --exclude-from="$EXCLUDE_FILE" \
    "$(basename $SOURCE_DIR)" \
    | ssh "$REMOTE_HOST" "cat > $REMOTE_DIR/$BACKUP_NAME"

# Limpiar archivo temporal
rm -f "$EXCLUDE_FILE"

# ============================================
# Verificar backup
# ============================================
REMOTE_SIZE=$(ssh "$REMOTE_HOST" "du -sh $REMOTE_DIR/$BACKUP_NAME | cut -f1")
echo -e "${GREEN}✅ Backup creado: $BACKUP_NAME ($REMOTE_SIZE)${NC}"

# ============================================
# Limpiar backups antiguos
# ============================================
echo -e "${YELLOW}🧹 Limpiando backups antiguos (>$KEEP_DAYS días)...${NC}"
ssh "$REMOTE_HOST" "find $REMOTE_DIR -name 'afs-backup-*.tar.gz' -mtime +$KEEP_DAYS -delete"

# ============================================
# Listar backups disponibles
# ============================================
echo ""
echo "📋 Backups disponibles en $REMOTE_HOST:"
ssh "$REMOTE_HOST" "ls -lh $REMOTE_DIR/*.tar.gz 2>/dev/null | tail -5" || echo "  (ninguno)"

echo ""
echo -e "${GREEN}✅ Backup completado exitosamente${NC}"