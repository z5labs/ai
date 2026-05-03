# Tables

Plausible columns inferred from the introspected counts; types are best-effort placeholders for ad-hoc-query orientation. Verify against `\d schema.table` before relying on them.

## analytics.cohort_assignments

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval(...) | PK |
| cohort_id | bigint | NO |  |  |
| user_id | bigint | NO |  |  |
| assigned_at | timestamptz | NO | now() |  |
| source | text | YES |  |  |
| metadata | jsonb | YES |  |  |

## analytics."UserSessions"

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval(...) | PK; referenced by `public.events.session_id`, `public.events.parent_session_id`, `crm.touchpoints.session_id` |
| user_id | bigint | NO |  |  |
| started_at | timestamptz | NO | now() |  |
| ended_at | timestamptz | YES |  |  |
| ip_address | inet | YES |  |  |
| user_agent | text | YES |  |  |
| referrer | text | YES |  |  |
| country | text | YES |  |  |
| device_type | text | YES |  |  |
| session_token | text | NO |  | unique |
| metadata | jsonb | YES |  |  |
| updated_at | timestamptz | NO | now() |  |

Identifier note: `UserSessions` is mixed-case, so SQL must double-quote it as `analytics."UserSessions"`. Bare `analytics.UserSessions` will be folded to lowercase by Postgres and fail.

## crm.touchpoints

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval(...) | PK |
| session_id | bigint | NO |  | FK → analytics."UserSessions".id |
| customer_id | bigint | NO |  |  |
| channel | text | NO |  |  |
| campaign_id | bigint | YES |  |  |
| occurred_at | timestamptz | NO | now() |  |
| value_cents | bigint | YES |  |  |
| properties | jsonb | YES |  |  |

## public.events

| column | type | null | default | notes |
|---|---|---|---|---|
| id | bigint | NO | nextval(...) | PK |
| session_id | bigint | NO |  | FK → analytics."UserSessions".id |
| parent_session_id | bigint | YES |  | FK → analytics."UserSessions".id |
| payload | jsonb | YES |  |  |
