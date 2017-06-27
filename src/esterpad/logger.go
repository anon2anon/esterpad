/*
Esterpad online collaborative editor
Copyright (C) 2016 Anon2Anon

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
	"fmt"
	"log"
	"os"
	"time"
)

const LOG_DEBUG = 0
const LOG_INFO = 1
const LOG_WARNING = 2
const LOG_ERROR = 3
const LOG_FATAL = 4

type Log struct {
	logFile  *os.File
	logLevel int
}

func LogInit(name string) Log {
	if Config == nil {
		ConfigRead("config.json")
	}
	l := Log{}

	logFile_local, err := os.OpenFile(Config["log"]["directory"].(string)+"/"+name+".log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		log.Fatal("log file open", err)
	}
	l.logLevel = int(Config["log"]["level"].(float64))
	l.logFile = logFile_local
	return l
}

func (l Log) Logf(level int, format string, v ...interface{}) {
	if level >= l.logLevel {
		t := time.Now()
		str := t.Format("2006/01/02 15:04:05") + " " + fmt.Sprintf(format, v...) + "\n"

		os.Stderr.WriteString(str)
		_, err := l.logFile.WriteString(str)
		if err != nil {
			log.Fatal("Log file write", err)
		}
	}
	if level == LOG_FATAL {
		os.Exit(1)
	}
}

func (l Log) Log(level int, v ...interface{}) {
	if level >= l.logLevel {
		t := time.Now()
		str := t.Format("2006/01/02 15:04:05") + " " + fmt.Sprintln(v...)
		os.Stderr.WriteString(str)

		_, err := l.logFile.WriteString(str)
		if err != nil {
			log.Fatal("Log file write", err)
		}
	}
	if level == LOG_FATAL {
		os.Exit(1)
	}
}
