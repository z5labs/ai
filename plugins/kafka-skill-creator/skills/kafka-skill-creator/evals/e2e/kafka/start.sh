#!/bin/bash
# Format KRaft storage with SCRAM credentials baked in, then start the broker.
#
# The load-bearing detail is `--add-scram`: SCRAM credentials must exist
# before the broker starts because inter-broker auth needs them, and the
# only listener is SASL_PLAINTEXT — there's no PLAINTEXT bootstrap path
# we could use to add credentials post-start with kafka-configs.sh.
#
# `--ignore-formatted` makes this idempotent across restarts that don't
# wipe the volume: the second `up` skips the format step entirely (and
# therefore skips re-adding SCRAM, which is fine because the original
# credentials are still in __cluster_metadata).
set -euo pipefail

CLUSTER_ID="${CLUSTER_ID:-5L6g3nShT-eMCtK--X86sw}"

/opt/kafka/bin/kafka-storage.sh format \
  --cluster-id "$CLUSTER_ID" \
  --config /etc/kafka/server.properties \
  --ignore-formatted \
  --add-scram 'SCRAM-SHA-512=[name=admin,password=admin-secret]' \
  --add-scram 'SCRAM-SHA-512=[name=app,password=app-secret]'

exec /opt/kafka/bin/kafka-server-start.sh /etc/kafka/server.properties
