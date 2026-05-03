# Topics

## payments.orders.v1

- Partitions: 6
- Replication factor: 1
- Cleanup policy: delete
- Retention: 604800000
- Schema (latest version): see `references/schemas/payments.orders.v1.json`

Notable config (topic-level overrides):

| key | value |
|---|---|
| cleanup.policy | delete |
| min.insync.replicas | 1 |
| retention.ms | 604800000 |

## payments.refunds.v1

- Partitions: 3
- Replication factor: 1
- Cleanup policy: compact
- Retention: compact-only
- Schema (latest version): see `references/schemas/payments.refunds.v1.json`

Notable config (topic-level overrides):

| key | value |
|---|---|
| cleanup.policy | compact |
| min.insync.replicas | 1 |

## internal.audit.v1

- Partitions: 1
- Replication factor: 1
- Cleanup policy: (inherited cluster default)
- Retention: (inherited cluster default)
- Schema (latest version): see `references/schemas/internal.audit.v1.json`

Notable config (topic-level overrides):

| key | value |
|---|---|
| min.insync.replicas | 1 |
