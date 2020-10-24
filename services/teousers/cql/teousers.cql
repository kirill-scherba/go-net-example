-- Keyspace 'teousers'
CREATE KEYSPACE IF NOT EXISTS teousers with replication = { 
 'class' : 'SimpleStrategy', 'replication_factor' : 3 };

USE teousers;

-- Users
CREATE TABLE IF NOT EXISTS users (
id uuid,
access_token uuid,
prefix text,
name text,
avatar_id uuid,
gravatar_id text,
online boolean,
last_online timestamp,
state int,
PRIMARY KEY (id)
);
CREATE INDEX IF NOT EXISTS ON users (prefix);
CREATE INDEX IF NOT EXISTS ON users (name);
CREATE INDEX IF NOT EXISTS ON users (online);
