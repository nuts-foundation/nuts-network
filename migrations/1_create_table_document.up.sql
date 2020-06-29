CREATE TABLE document
(
    hash             CHAR(40) PRIMARY KEY,
    consistency_hash CHAR(40),
    type             VARCHAR(100) NOT NULL,
    timestamp        BIGINT       NOT NULL,
    contents         BLOB
);

-- We sort on timestamp, hash so in theory an index should speed up things
-- TODO: Test this!
CREATE INDEX idx_timestamp_hash ON document (timestamp, hash);
CREATE INDEX idx_consistency_hash ON document (consistency_hash);

CREATE TABLE stats
(
    key VARCHAR(40) PRIMARY KEY,
    value VARCHAR(100)
);
CREATE TRIGGER update_stats_contents_size UPDATE OF contents ON document
BEGIN
    UPDATE stats SET value = CAST((CAST(value AS INTEGER) + IFNULL(LENGTH(new.contents), 0) - IFNULL(LENGTH(old.contents), 0)) AS VARCHAR) WHERE key = 'contents-size';
END;
INSERT INTO stats VALUES ('contents-size', 0);