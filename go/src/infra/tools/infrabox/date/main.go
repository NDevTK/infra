// +build android

// Command line utility used to set a device's system time. Mainly used for
// android devices. Reimplemented here because... android.
package main

/*
#cgo LDFLAGS: -landroid

#include <sys/time.h>
*/
import "C"

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/luci/luci-go/common/runtime/paniccatcher"
)

func realMain() int {
	seconds := flag.Int("s", -1, "Number of seconds since the epoch to set the system time to.")
	flag.Parse()

	if *seconds == -1 {
		t := time.Now().UTC()
		fmt.Fprintln(os.Stdout, t.Format("2006-01-02 15:04:05 UTC"))
	} else if *seconds > 0 {
		tv := C.struct_timeval{}
		tv.tv_sec = C.time_t(*seconds)
		tz := C.struct_timezone{}
		C.settimeofday(&tv, &tz)
	} else {
		fmt.Fprintf(os.Stderr, "Invalid arguments.\n")
		return 1
	}

	return 0
}

func main() {
	paniccatcher.Do(func() {
		os.Exit(realMain())
	}, func(p *paniccatcher.Panic) {
		fmt.Fprintln(os.Stderr, p.Reason)
		fmt.Fprintln(os.Stderr, p.Stack)
		os.Exit(1)
	})
}
