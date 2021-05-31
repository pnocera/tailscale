// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Program speedtest provides the speedtest command. The reason to keep it separate from
// the normal tailscale cli is because it is not yet ready to go in the tailscale binary.
// It will be included in the tailscale cli after it has been added to tailscaled.
// Example usage for client command: go run cmd/speedtest -d -host 127.0.0.1 -port 8080 -t 5s
// This will connect to the server on 127.0.0.1:8080 and start a 5 second download speedtest.
// Example usage for server command: go run cmd/speedtest -s -host :8080
// This will start a speedtest server on port 8080.
// at a time.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"text/tabwriter"

	"tailscale.com/net/speedtest"

	"github.com/peterbourgon/ff/v2/ffcli"
)

// Runs the speedtest command as a commandline program
func main() {
	args := os.Args[1:]
	if err := speedtestCmd.Parse(args); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}

	err := speedtestCmd.Run(context.Background())
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println(speedtestCmd.ShortUsage)
		os.Exit(2)
	}
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}
}

// speedtestCmd is the root command. It runs either the server and client depending on the
// flags passed to it.
var speedtestCmd = &ffcli.Command{
	Name:       "speedtest",
	ShortUsage: "speedtest [-s] [-max-conns <max connections>] [-t <test duration>]",
	ShortHelp:  "Run a speed test",
	FlagSet: (func() *flag.FlagSet {
		fs := flag.NewFlagSet("speedtest", flag.ExitOnError)
		fs.StringVar(&speedtestArgs.host, "host", "", "host to connect to or listen on")
		fs.IntVar(&speedtestArgs.port, "port", 20333, "port to connect to or listen on")
		fs.DurationVar(&speedtestArgs.testDuration, "t", speedtest.DefaultDuration, "duration of the speed test")
		fs.BoolVar(&speedtestArgs.runServer, "s", false, "run a speedtest server")
		fs.BoolVar(&speedtestArgs.downloadTest, "d", false, "conduct a download test instead of an upload test")
		return fs
	})(),
	Exec: runSpeedtest,
}

var speedtestArgs struct {
	port         int
	host         string
	testDuration time.Duration
	runServer    bool
	downloadTest bool
}

func runSpeedtest(ctx context.Context, args []string) error {

	host := net.JoinHostPort(speedtestArgs.host, strconv.Itoa(speedtestArgs.port))
	if speedtestArgs.runServer {
		listener, err := net.Listen("tcp", host)
		if err != nil {
			return err
		}

		// If the user provides a 0 port, a random available port will be chosen,
		// so we need to identify which one was chosen, to display to the user.
		port := listener.Addr().(*net.TCPAddr).Port
		fmt.Println("listening on port", port)

		return speedtest.Serve(listener)
	}

	if speedtestArgs.host == "" {
		return errors.New("both host and port must be given")
	}
	// Ensure the duration is within the allowed range
	if speedtestArgs.testDuration < speedtest.MinDuration || speedtestArgs.testDuration > speedtest.MaxDuration {
		return errors.New(fmt.Sprintf("test duration must be within %.0fs and %.0fs.\n", speedtest.MinDuration.Seconds(), speedtest.MaxDuration.Seconds()))
	}

	if speedtestArgs.downloadTest {
		fmt.Printf("Starting a download test with %s\n", speedtestArgs.host)
	} else {
		fmt.Printf("Starting an upload test with %s\n", speedtestArgs.host)
	}
	results, err := speedtest.RunClient(speedtestArgs.downloadTest, speedtestArgs.testDuration, host)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 12, 0, 0, ' ', tabwriter.TabIndent)
	fmt.Println("Results:")
	fmt.Fprintln(w, "Interval\t\tTransfer\t\tBandwidth\t\t")
	for _, r := range results {
		if r.Total {
			fmt.Fprintln(w, "-------------------------------------------------------------------------")
		}
		fmt.Fprintf(w, "%.2f-%.2f\tsec\t%.4f\tMBytes\t%.4f\tMbits/sec\t\n", r.IntervalStart.Seconds(), r.IntervalEnd.Seconds(), r.MegaBytes(), r.MBitsPerSecond())
	}
	w.Flush()
	return nil
}
