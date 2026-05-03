# Source this (don't execute it) to export the per-context env vars the
# kafka-skill-creator skill expects when pointed at this fixture.
#
#   source ./env.sh
#
# These match docker-compose.yml's SCRAM credentials and Karapace authfile.
# They are intentionally not secrets — the fixture is throwaway and the
# values are checked in.

export CONTEXTS_DEV_BROKERS=localhost:9092
export CONTEXTS_DEV_SASL_USERNAME=app
export CONTEXTS_DEV_SASL_PASSWORD=app-secret
export CONTEXTS_DEV_SCHEMAREGISTRY_URL=http://localhost:8081
export CONTEXTS_DEV_SCHEMAREGISTRY_USERNAME=sruser
export CONTEXTS_DEV_SCHEMAREGISTRY_PASSWORD=sr-secret

# The kafkactl + curl containers used by introspect.sh and the SR fetch
# snippet need to reach localhost:9092 / localhost:8081 on the host.
# --network=host is the documented Linux path; macOS support is tracked
# in #78.
export KAFKA_DOCKER_ARGS=--network=host

# Pin the container runtime to whichever one currently owns the fixture.
# up.sh exports KAFKA_CONTAINER_RUNTIME for its own subshell, but that
# value dies with the script — by the time the user `source`s this file
# in a separate shell, it's gone. Without this, subsequent introspect.sh
# / generated-wrapper calls would auto-detect, prefer docker on mixed
# hosts, and talk to the wrong engine.
#
# Same detection pattern down.sh uses: ask each available runtime whether
# it currently holds the fixture's named container. If neither does, the
# fixture isn't up — leave the var unset so the wrappers' default
# auto-detect handles a fresh runtime.
for _r in docker podman; do
  command -v "$_r" >/dev/null 2>&1 || continue
  if "$_r" ps -a --format '{{.Names}}' 2>/dev/null \
      | grep -qx kafka-skill-creator-e2e-kafka; then
    export KAFKA_CONTAINER_RUNTIME="$_r"
    break
  fi
done
unset _r
