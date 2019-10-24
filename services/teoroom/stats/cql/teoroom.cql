/* Keyspace 'teoroom' */
CREATE KEYSPACE IF NOT EXISTS teoroom WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor' : 3
};

USE teoroom;

/* Room state */
CREATE TABLE IF NOT EXISTS rooms (
   id uuid,
   room_num int,
   created timestamp,
   started timestamp,
   closed timestamp,
   stopped timestamp,
   state int,
   PRIMARY KEY (id)
 );
 CREATE INDEX IF NOT EXISTS ON rooms (room_num);
 CREATE INDEX IF NOT EXISTS ON rooms (created);
 CREATE INDEX IF NOT EXISTS ON rooms (started);
 CREATE INDEX IF NOT EXISTS ON rooms (closed);
 CREATE INDEX IF NOT EXISTS ON rooms (stopped);
 CREATE INDEX IF NOT EXISTS ON rooms (state);
 
 /* Room client state */
 CREATE TABLE IF NOT EXISTS clients (
   room_id uuid,
   id uuid,
   added timestamp,
   loadded timestamp,
   started timestamp,
   leave timestamp,
   disconnected timestamp,
   game_stat blob,
   PRIMARY KEY (room_id, id)
 );
 CREATE INDEX IF NOT EXISTS ON clients (added);
 CREATE INDEX IF NOT EXISTS ON clients (loadded);
 CREATE INDEX IF NOT EXISTS ON clients (started);
 CREATE INDEX IF NOT EXISTS ON clients (leave);
 CREATE INDEX IF NOT EXISTS ON clients (disconnected);
