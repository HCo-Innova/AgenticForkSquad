#!/bin/bash
source ~/venv/bin/activate
export ACCESS_TOKEN=$(/usr/bin/env python3 ./scripts/generate_token.py)
echo "$(date '+%F %T') | Token actualizado" >> /var/log/token_refresh.log
