// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Redirect std out module of teoapi package.

package teoapi

import (
	"os"
	"syscall"
)

// Stdout module receiver
type Stdout struct {
	stdout *os.File
	stderr *os.File
}

// NewStdout creates stdout module receiver
func NewStdout() (s *Stdout) {
	s = &Stdout{stdout: os.Stdout, stderr: os.Stderr}
	return
}

// Redirect standart output to file
func (s *Stdout) Redirect() {
	os.Stdout = os.NewFile(uintptr(syscall.Stdout), "/dev/nill")
	os.Stderr = os.NewFile(uintptr(syscall.Stderr), "/dev/nill")
}

// Restore standart output
func (s *Stdout) Restore() {
	os.Stdout = s.stdout
	os.Stderr = s.stderr
}
