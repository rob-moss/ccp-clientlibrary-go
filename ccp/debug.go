/*Copyright (c) 2019 Cisco and/or its affiliates.

This software is licensed to you under the terms of the Cisco Sample
Code License, Version 1.0 (the "License"). You may obtain a copy of the
License at

               https://developer.cisco.com/docs/licenses

All use of the material herein must be in accordance with the terms of
the License. All rights not expressly granted by the License are
reserved. Unless required by applicable law or agreed to separately in
writing, software distributed under the License is distributed on an "AS
IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
or implied.*/

package ccp

import (
	"fmt"
)

// debug levels
// var debuglvl int64 = int64(0) // debug off
// var debuglvl int64 = int64(1) // 1 basic debugging, errors and warnings
var debuglvl int64 = int64(2) // 2 medium debugging, above + some data
// var debuglvl int64 = int64(3) // 3 high debugging, above + all json input/output and structs

// Debug messages
func Debug(level int64, errmsg string) {

	if level <= debuglvl {
		fmt.Println("Debug: " + errmsg)
	}
}
