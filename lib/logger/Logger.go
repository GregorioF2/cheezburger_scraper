package logger

import (
	"fmt"
	config "propper/configs"
	"time"
)

func init() {
	if config.DEBUG {
		fmt.Println("DEBUG MODE IS ON")
	}
}

func Log(i ...interface{}) {
	if !config.DEBUG {
		return
	}
	now := time.Now().UTC().Format("15:04:05")
	fmt.Printf("%s :: ", now)
	fmt.Println(i...)
}
