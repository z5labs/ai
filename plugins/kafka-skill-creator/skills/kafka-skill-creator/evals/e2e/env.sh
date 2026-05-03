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
