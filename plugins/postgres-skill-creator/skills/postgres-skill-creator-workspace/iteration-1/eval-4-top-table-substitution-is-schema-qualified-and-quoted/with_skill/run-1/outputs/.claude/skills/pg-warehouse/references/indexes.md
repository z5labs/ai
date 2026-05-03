# Indexes

One line per index, schema-qualified. Use this to judge whether a WHERE clause hits an index before issuing it.

analytics."UserSessions": CREATE UNIQUE INDEX "UserSessions_pkey" ON analytics."UserSessions" USING btree (id)
analytics."UserSessions": CREATE UNIQUE INDEX "UserSessions_session_token_key" ON analytics."UserSessions" USING btree (session_token)
analytics."UserSessions": CREATE INDEX "UserSessions_user_id_idx" ON analytics."UserSessions" USING btree (user_id)
analytics.cohort_assignments: CREATE UNIQUE INDEX cohort_assignments_pkey ON analytics.cohort_assignments USING btree (id)
analytics.cohort_assignments: CREATE INDEX cohort_assignments_user_id_idx ON analytics.cohort_assignments USING btree (user_id)
crm.touchpoints: CREATE UNIQUE INDEX touchpoints_pkey ON crm.touchpoints USING btree (id)
crm.touchpoints: CREATE INDEX touchpoints_session_id_idx ON crm.touchpoints USING btree (session_id)
public.events: CREATE UNIQUE INDEX events_pkey ON public.events USING btree (id)
public.events: CREATE INDEX events_session_id_idx ON public.events USING btree (session_id)
public.events: CREATE INDEX events_parent_session_id_idx ON public.events USING btree (parent_session_id)
