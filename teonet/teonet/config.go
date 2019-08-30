// Copyright 2019 teonet-go authors.  All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

// Teonet config module:
//
// Save restore teonet parameters from configuration file

package teonet

// configNew initialize configuration receiver and read config file
func (param *Parameters) ReadConfig() {
	param.read()
}

func (param *Parameters) read() {
	// jsonFile, err := os.Open("users.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println("Successfully Opened users.json")
	// defer jsonFile.Close()
	// return
}

func (param *Parameters) write() {

}
