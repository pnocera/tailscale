// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package speedtest

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

// Serve starts up the server on a given host and port pair. It starts to listen for
// connections and handles each one in a goroutine. Because it runs in an infinite loop,
// this function only returns if any of the speedtests return with errors, or if the
// listener is closed.
func Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if opErr, ok := err.(*net.OpError); ok {
			if opErr.Err.Error() == listenerClosedErr {
				// listener closed
				return nil
			}
		}
		if err != nil {
			return err
		}
		err = handleConnection(conn)
		if err != nil {
			return err
		}
	}
}

// handleConnection handles the initial exchange between the server and the client.
// It reads the testconfig message into a testConfig struct. If any errors occur with
// the testconfig (specifically, if there is a version mismatch), it will return those
// errors to the client with a testConfigResponse. After the exchange, it will start
// the speed test.
func handleConnection(conn net.Conn) error {
	defer conn.Close()
	var config testConfig

	decoder := json.NewDecoder(conn)
	err := decoder.Decode(&config)

	encoder := json.NewEncoder(conn)
	// Both return and encode errors that were thrown before the test has started
	if err != nil {
		encoder.Encode(testConfigResponse{Error: err.Error()})
		return err
	}

	if config.Version != version {
		err = fmt.Errorf("version mismatch! Server is version %d, client is version %d", version, config.Version)
		encoder.Encode(testConfigResponse{Error: err.Error()})
		return err
	}

	// Start the test
	encoder.Encode(testConfigResponse{Error: ""})
	// when the client does download, the server does upload
	if config.DownloadTest {
		_, err = runUpload(conn, config)
	} else {
		_, err = runDownload(conn, config)
	}
	return err
}

// runUpload runs the server side of the speed test. For the given amount of time,
// it sends data in 32 kilobyte blocks. When time's up the function returns
// and the connection is closed. This function returns an error if the write fails,
// as well as a slice of results that contains the result of the test.
func runUpload(conn net.Conn, config testConfig) ([]Result, error) {
	BufData := make([]byte, blockSize)
	intervalBytes := 0
	totalBytes := 0

	var lastCalculated time.Time
	var currentTime time.Time
	var startTime time.Time
	var results []Result

	for {
		// Randomize data
		_, err := rand.Read(BufData)
		if err != nil {
			continue
		}

		n, err := conn.Write(BufData)
		currentTime = time.Now()
		intervalBytes += n
		if err != nil {
			// If the write failed, there is most likely something wrong with the connection.
			return nil, fmt.Errorf("server: connection closed unexpectedly: %w", err)
		}
		if startTime.IsZero() {
			startTime = time.Now()
			lastCalculated = time.Now()
		}

		if currentTime.After(lastCalculated.Add(increment)) {
			intervalStart := lastCalculated.Sub(startTime)
			intervalEnd := currentTime.Sub(startTime)
			if (intervalEnd - intervalStart) > minInterval {
				results = append(results, Result{Bytes: intervalBytes, IntervalStart: intervalStart, IntervalEnd: intervalEnd, Total: false})
			}
			lastCalculated = lastCalculated.Add(increment)
			totalBytes += intervalBytes
			intervalBytes = 0
		}

		if time.Since(startTime) > config.TestDuration {
			break
		}
	}

	if currentTime.Sub(startTime) > minInterval {
		intervalEnd := currentTime.Sub(startTime)
		results = append(results, Result{Bytes: totalBytes, IntervalStart: 0, IntervalEnd: intervalEnd, Total: true})
	}
	return results, nil
}
