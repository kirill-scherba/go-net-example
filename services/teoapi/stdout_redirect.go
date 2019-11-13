// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Redirect std out module of teoapi package.

package teoapi

import (
	"os"
	"syscall"
)

// Stdout
type Stdout struct {
	stdout *os.File
	stderr *os.File
}

// NewStdout creates stdout module receiver
func NewStdout() (s *Stdout) {
	s = &Stdout{stdout: os.Stdout, stderr: os.Stderr}
	return
}

// redirect standart output to file
func (stdout *Stdout) redirect() {
	// f, _ := os.OpenFile("/tmp/teocli-termloop",
	// 	os.O_WRONLY|os.O_CREATE|os.O_SYNC, 0755)
	os.Stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/nill")
	os.Stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/nill")
}

// restory standart output
func (stdout *Stdout) restory() {
	os.Stdout = stdout
	os.Stderr = stderr
}
