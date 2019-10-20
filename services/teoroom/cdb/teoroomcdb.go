package cdb

// Rooms is the teoroomcdb data structure and methods receiver
type Rooms struct {
	*db
	*Process
	TeoConnector
}

// Connect to the cql cluster and create teoroomcdb receiver.
// First parameter is keyspace, next parameters is hosts name (usualy it should
// be 3 hosts - 3 ScyllaDB nodes)
func Connect(con TeoConnector, hosts ...string) (r *Rooms, err error) {
	r = &Rooms{TeoConnector: con}
	r.db, err = newDb(hosts...)
	r.Process = &Process{r}
	return
}

// Close closes cql connection and destroy teoregistry receiver
func (r *Rooms) Close() {
	r.db.close()
}
