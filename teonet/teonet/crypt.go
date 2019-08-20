package teonet

//// CGO definition (don't delay or edit it):
//#cgo LDFLAGS: -lcrypto
//#include "crypt.h"
import "C"
import (
	"unsafe"

	"github.com/kirill-scherba/net-example-go/teolog/teolog"
)

// Teonet crypt module

type crypt struct {
	kcr *C.ksnCryptClass
}

// initialize Init crypt module
func (teo *Teonet) cryptoNew(key string) *crypt {
	ckey := append([]byte(key), 0)
	ckeyPtr := (*C.char)(unsafe.Pointer(&ckey[0]))
	cry := &crypt{kcr: C.ksnCryptInit(ckeyPtr)}
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
	if cry.kcr == nil {
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
func (cry *crypt) decrypt(packet []byte, key string) []byte {
	if cry.kcr == nil {
		return packet
	}
	var decryptLen C.size_t
	packetPtr := unsafe.Pointer(&packet[0])
	C.ksnDecryptPackage(cry.kcr, packetPtr, C.size_t(len(packet)), &decryptLen)
	if decryptLen > 0 {
		packet = packet[2 : decryptLen+2]
		teolog.DebugVvf(MODULE, "decripted to %d bytes packet, channel key: %s\n", decryptLen, key)
	} else {
		teolog.DebugVvf(MODULE, "can't decript %d bytes packet (try to use without decrypt), channel key: %s\n", len(packet), key)
	}
	return packet
}
