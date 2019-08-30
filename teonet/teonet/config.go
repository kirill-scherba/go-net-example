// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet config module:
//
// Save restore teonet parameters from configuration file

package teonet

type config struct {
	teo *Teonet
}

// configNew initialize config receiver
func (teo *Teonet) configNew() (conf *config) {
	conf = &config{teo: teo}
	return
}

func (conf *config) read() (param *Parameters) {
	return
}

func (conf *config) write(param *Parameters) {

}
