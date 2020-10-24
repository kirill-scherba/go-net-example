--  Keyspace 'teoregistry'
CREATE KEYSPACE IF NOT EXISTS teoregistry WITH replication = { 
'class': 'SimpleStrategy', 'replication_factor' : 3 };
USE teoregistry;

--  Table 'applications': Teonet applications (services) description
CREATE TABLE IF NOT EXISTS applications(
uuid        TIMEUUID,
name        TEXT,
version     TEXT,
descr       TEXT,
author      TEXT,
license     TEXT,
goget       TEXT,
git         TEXT,
PRIMARY KEY(uuid)
);
CREATE INDEX IF NOT EXISTS ON applications (name);

-- Table 'commands': Teonet applications commands description,
-- type values:  0 - input; 1 - input/output (same parameters); 2 - output
CREATE TABLE IF NOT EXISTS commands(
app_id       TIMEUUID,
cmd          INT,
type         TINYINT,
descr        TEXT,
txt_f        BOOLEAN,
txt_num      TINYINT,
txt_descr    TEXT,
jsonf        BOOLEAN,
json         TEXT,
binary_f     BOOLEAN,
binary_descr TEXT,
PRIMARY KEY(app_id, cmd, type)
);
