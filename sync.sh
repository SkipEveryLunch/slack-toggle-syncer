#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

SLACK_BOT_TOKEN="$(security find-generic-password -a "$USER" -s slack-toggle-syncer.slack_bot_token -w)"
SLACK_CHANNEL_ID="$(security find-generic-password -a "$USER" -s slack-toggle-syncer.slack_channel_id -w)"
TOGGL_API_TOKEN="$(security find-generic-password -a "$USER" -s slack-toggle-syncer.toggl_api_token -w)"
TOGGL_WORKSPACE_ID="$(security find-generic-password -a "$USER" -s slack-toggle-syncer.toggl_workspace_id -w)"

exec env \
  LOG_LEVEL=info \
  SLACK_BOT_TOKEN="$SLACK_BOT_TOKEN" \
  SLACK_CHANNEL_ID="$SLACK_CHANNEL_ID" \
  TOGGL_API_TOKEN="$TOGGL_API_TOKEN" \
  TOGGL_WORKSPACE_ID="$TOGGL_WORKSPACE_ID" \
  go run .