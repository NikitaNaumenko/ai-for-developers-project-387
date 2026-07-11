CREATE UNIQUE INDEX event_types_title_description_unique_idx
    ON event_types (lower(btrim(title)), lower(btrim(description)));
