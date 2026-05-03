# Relationships

Foreign keys in arrow form. `schema.table.column → ref_schema.ref_table.ref_column`.

public.events.session_id → analytics."UserSessions".id
public.events.parent_session_id → analytics."UserSessions".id
crm.touchpoints.session_id → analytics."UserSessions".id
