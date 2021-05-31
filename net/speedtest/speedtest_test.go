// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package speedtest

import (
	"fmt"
	"net"
	"testing"
)

func TestDownload(t *testing.T) {
	// start a lisenter and find the port where the server will be listening.
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })

	serverIP := l.Addr().String()
	t.Log("server IP found:", serverIP)

	type state struct {
		err error
	}

	stateChan := make(chan state, 1)

	go func() {
		err := Serve(l)
		stateChan <- state{err: err}
	}()

	// ensure that the test returns an appropriate number of Result structs
	expectedLen := int(DefaultDuration.Seconds()) + 1
	downloadTest := true
	// conduct a download test
	results, err := RunClient(downloadTest, DefaultDuration, serverIP)
	if err != nil {
		t.Error("download test failed:", err)
	} else {
		if len(results) < expectedLen {
			t.Errorf("download results: expected length: %d, actual length: %d", expectedLen, len(results))
		}
		for _, result := range results {
			t.Log(resultToString(result))
		}
	}

	// conduct an upload test
	downloadTest = false
	results, err = RunClient(downloadTest, DefaultDuration, serverIP)
	if err != nil {
		t.Error("upload test failed:", err)
	} else {
		if len(results) < expectedLen {
			t.Errorf("upload results: expected length: %d, actual length: %d", expectedLen, len(results))
		}
		for _, result := range results {
			t.Log(resultToString(result))
		}
	}

	// causes the server goroutine to finish
	l.Close()

	testState := <-stateChan
	if testState.err != nil {
		t.Error("server error:", err)
	}
}

func resultToString(r Result) string {
	return fmt.Sprintf("{ Megabytes: %.2f, Start: %.1f, End: %.1f, Total: %t }", r.MegaBytes(), r.IntervalStart.Seconds(), r.IntervalEnd.Seconds(), r.Total)
}
