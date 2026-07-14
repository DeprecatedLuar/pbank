#!/usr/bin/env bash
# Monthly cashflow report fan-out.
#
# A routine fires once, with no user/session bound to it. To reach "every user"
# we iterate the agent's identities ourselves: for each identity linked to a
# delivery channel, run a one-shot `agentctl chat` scoped to that user (--user)
# and deliver the AI's summary straight to their channel (--deliver). The chat
# generates a real completion, so the pbank history tool runs against that
# user's own PBANK_HOME (scoped by {{$user}}) and the summary is per-user.
#
# All tunables (delivery channel, prompt) live in
# config/routines/monthly_report/ — this script holds only the fan-out logic.
# Schedule lives in routines/monthly_report/monthly_report.toml.
#
# Inputs (injected by the routine as env vars):
#   ROUTINE_AGENTPATH - absolute agent folder
#   ROUTINE_REFDATE   - fire date (YYYY-MM-DD), always the 1st

set -euo pipefail

AGENT="${ROUTINE_AGENTPATH:?agentpath not provided}"
REF="${ROUTINE_REFDATE:?refdate not provided}"
CONTACTS="$AGENT/.data/contacts.toml"
CONF="$AGENT/config/routines/monthly_report/monthly_report.conf"

[ -f "$CONF" ] || { echo "monthly_report: missing config $CONF" >&2; exit 1; }
# shellcheck disable=SC1090
source "$CONF"
: "${DELIVER_TO:?DELIVER_TO not set in $CONF}"
: "${PROMPT_FILE_PATH:?PROMPT_FILE_PATH not set in $CONF}"

PROMPT_PATH="$AGENT/$PROMPT_FILE_PATH"
[ -f "$PROMPT_PATH" ] || { echo "monthly_report: missing prompt file $PROMPT_PATH" >&2; exit 1; }
PROMPT="$(cat "$PROMPT_PATH")"

# Previous calendar month, derived from the fire date (the 1st).
SINCE="$(date -d "$REF -1 month" +%Y-%m-01)"
LABEL="$(date -d "$SINCE" +'%B %Y')"

# Substitute the prompt's placeholders.
PROMPT="${PROMPT//\{LABEL\}/$LABEL}"
PROMPT="${PROMPT//\{SINCE\}/$SINCE}"

# The interface prefix that marks a deliverable contact in contacts.toml, taken
# from the first channel in DELIVER_TO (e.g. "telegram" -> "telegram:").
CHANNEL="${DELIVER_TO%%,*}"

# Emit the id of every [[identity]] whose contacts list includes the channel.
channel_identities() {
  awk -v chan="${CHANNEL}:" '
    /^\[\[identity\]\]/ { curid=""; inident=1; next }
    /^\[\[/ && $0 !~ /identity/ { inident=0 }
    inident && /^[[:space:]]*id[[:space:]]*=/ {
      if (match($0, /"[^"]+"/)) curid = substr($0, RSTART+1, RLENGTH-2)
    }
    inident && /^[[:space:]]*contacts[[:space:]]*=/ && index($0, chan) {
      if (curid != "") print curid
    }
  ' "$CONTACTS"
}

sent=0
while IFS= read -r uid; do
  [ -z "$uid" ] && continue
  echo "monthly_report: sending ${LABEL} summary to '$uid' via ${DELIVER_TO}"
  if agentctl chat -m "$PROMPT" --user "$uid" --deliver "$DELIVER_TO" -a "$AGENT"; then
    sent=$((sent + 1))
  else
    echo "monthly_report: delivery failed for '$uid'" >&2
  fi
done < <(channel_identities)

echo "monthly_report: done — delivered to $sent user(s) for $LABEL"
