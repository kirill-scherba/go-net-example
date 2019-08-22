package teonet

// Teonet L0 server module

type l0 struct {
	teo *Teonet
}

// l0New initialize l0 module
func (teo *Teonet) l0tNew(key string) *l0 {
	l0 := &l0{teo: teo}
	return l0
}

// destroy destroys l0 module
func (l0 *l0) destroy() {

}
