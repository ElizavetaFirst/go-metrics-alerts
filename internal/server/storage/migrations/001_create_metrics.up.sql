CREATE TABLE IF NOT EXISTS metrics (
                                       name text NOT NULL,
                                       type text NOT NULL,
                                       value double precision,
                                       delta bigint,
                                       PRIMARY KEY (name, type)
    );
