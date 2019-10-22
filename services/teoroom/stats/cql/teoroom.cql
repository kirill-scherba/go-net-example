/* Keyspace 'teoroom' */
CREATE KEYSPACE IF NOT EXISTS teoroom WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor' : 3
};

USE teoroom;

CREATE TABLE IF NOT EXISTS  rooms (
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
 