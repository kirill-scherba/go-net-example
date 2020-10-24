// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet crypt module.

package teonet

// Install this C dependences to use crypto:
//   sudo apt-get install -y libssl-dev

// #cgo LDFLAGS: -lcrypto
// #include "crypt.h"
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/kirill-scherba/teonet-go/teolog/teolog"
)

type crypt struct {
	teo *Teonet
	kcr *C.ksnCryptClass
}

// cryptNew initialize crypt module
func (teo *Teonet) cryptNew(key string) *crypt {
	ckey := append([]byte(key), 0)
	ckeyPtr := (*C.char)(unsafe.Pointer(&ckey[0]))
	cry := &crypt{teo: teo, kcr: C.ksnCryptInit(ckeyPtr)}
	return cry
}

// destroy Destroy crypt module
func (cry *crypt) destroy() {
	if cry.kcr != nil {
		C.ksnCryptDestroy(cry.kcr)
		cry.kcr = nil
	}
}

// encryptp Encryptp teonet packet
func (cry *crypt) encrypt(packet []byte) []byte {
	if cry.kcr == nil || cry.teo.param.DisallowEncrypt {
		return packet
	}
	buf := make([]byte, len(packet)+int(C.ksnCryptGetBlockSize(cry.kcr))+C.sizeof_size_t)
	var encryptLen C.size_t
	bufPtr := unsafe.Pointer(&buf[0])
	packetPtr := unsafe.Pointer(&packet[0])
	C.ksnEncryptPackage(cry.kcr, packetPtr, C.size_t(len(packet)), bufPtr, &encryptLen)
	buf = buf[:encryptLen]
	return buf
}

// packet Decrypt teonet packet
func (cry *crypt) decrypt(packet []byte, key string) ([]byte, error) {
	if cry.kcr == nil {
		return packet, errors.New("crypt module does not initialized")
	}

	errCantDecript := func() (err error) {
		err = fmt.Errorf("can't decript %d bytes packet (try to use "+
			"without decrypt), channel key: %s", len(packet), key)
		teolog.DebugVv(MODULE, err.Error())
		return
	}

	// Empty packet
	if packet == nil || len(packet) == 0 {
		return packet, errCantDecript()
	}

	var err error
	var decryptLen C.size_t
	packetPtr := unsafe.Pointer(&packet[0])
	C.ksnDecryptPackage(cry.kcr, packetPtr, C.size_t(len(packet)), &decryptLen)
	if decryptLen > 0 {
		packet = packet[2 : decryptLen+2]
		teolog.DebugVvf(MODULE, "decripted to %d bytes packet, channel key: %s\n",
			decryptLen, key)
	} else {
		err = errCantDecript()
	}
	return packet, err
}
