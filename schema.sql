create table if not exists flights (
    -- the flight number/designator e.g, AA123 or N1234C
    flight_designator text,
    -- when the flight was observed
    seen_time timestamp with time zone,
    -- composite primary key based on both the flight designator and the time it
    -- was seen so that duplicates cannot be inserted
    primary key (flight_designator, seen_time)
);
