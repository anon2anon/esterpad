/*
Esterpad online collaborative editor
Copyright (C) 2017 Anon2Anon

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/

package esterpad

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

var Config map[string]map[string]interface{}

func ConfigRead(fname string) {
	dat, err := ioutil.ReadFile(fname)
	if err != nil {
		panic(err)
	}

	Config = make(map[string]map[string]interface{})

	configInterface := interface{}(nil)
	json.Unmarshal([]byte(string(dat)), &configInterface)
	for key, value := range configInterface.(map[string]interface{}) {
		Config[key] = value.(map[string]interface{})
	}
	log.Println("Config has been read")
}
