// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Redirect std out module of teoapi package.

package teoapi

import (
	"os"
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
	null, _ := os.Open(os.DevNull)
	os.Stdout = null
	os.Stderr = null
}

// Restore standart output
func (s *Stdout) Restore() {
	os.Stdout = s.stdout
	os.Stderr = s.stderr
}
