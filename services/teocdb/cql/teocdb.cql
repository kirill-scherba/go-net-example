-- Keyspace 'teocdb'
create keyspace IF NOT EXISTS teocdb with replication = { 
'class' : 'SimpleStrategy', 'replication_factor' : 3 };
create table IF NOT EXISTS teocdb.map(key text, data blob, PRIMARY KEY(key));
create table IF NOT EXISTS teocdb.ids(id_name text, next_id int, PRIMARY KEY(id_name));
