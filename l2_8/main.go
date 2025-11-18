package main

import (
	"fmt"
	"os"
	"time"

	"github.com/beevik/ntp"
)

const ntpServer = "pool.ntp.org"

func main() {
	currentTime, err := ntp.Time(ntpServer)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(currentTime.Format(time.RFC3339))
}
