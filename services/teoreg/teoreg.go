// Copyright 2019 Teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Package teoreg (teo-reg) is the Teonet registrar service package.
//
// Dependences (communicate with): teousers (teo-cdb with users package).
//
// This module process registrar comands:
//
// Receive requests. The request is teonet client login name. The login
// name contains prefix and access key: tg001-756-44-33, where: tg001 is prefix,
// and 756-44-33 is access key. Prefix may contain '-new' ending, then user
// shoud be added: tg001-new-...
//
// Process received requests with teousers functions, which connect to teo-cdb
// and check and add users to teonet users database.
//
package teoreg

