// Copyright (c) 2021 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package speedtest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"time"
)

// RunClient dials the given address and starts a speedtest.
// It returns any errors that come up in the tests.
// If there are no errors in the test, it returns a slice of results.
func RunClient(downloadTest bool, duration time.Duration, host string) ([]Result, error) {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return nil, err
	}

	config := testConfig{TestDuration: duration, Version: version, DownloadTest: downloadTest}

	defer conn.Close()
	encoder := json.NewEncoder(conn)

	if err = encoder.Encode(config); err != nil {
		return nil, err
	}

	var response testConfigResponse
	decoder := json.NewDecoder(conn)
	if err = decoder.Decode(&response); err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, errors.New(response.Error)
	}
	if config.DownloadTest {
		return runDownload(conn, config)
	} else {
		return runUpload(conn, config)
	}
}

// TODO include code to detect whether the code is direct vs DERP

// runDownload handles the entire download speed test.
// It has a loop that breaks if the connection receives an EOF.
// It keeps track of the amount of data downloaded and stores that data in
// a result slice.
func runDownload(conn net.Conn, config testConfig) ([]Result, error) {
	bufferData := make([]byte, blockSize)

	intervalBytes := 0
	totalBytes := 0

	var currentTime time.Time
	var results []Result

	downloadBegin := time.Now()
	lastCalculated := downloadBegin

	conn.SetReadDeadline(time.Now().Add(config.TestDuration).Add(5 * time.Second))
Receive:
	for {
		n, err := io.ReadFull(conn, bufferData)

		currentTime = time.Now()
		intervalBytes += n
		switch err {
		case io.EOF, io.ErrUnexpectedEOF:
			break Receive
		case nil:
			// successful read
		default:
			return nil, fmt.Errorf("unexpected error has occured: %w", err)
		}

		// checks if the current time is more or equal to the lastCalculated time plus the increment
		if currentTime.After(lastCalculated.Add(increment)) {
			intervalStart := lastCalculated.Sub(downloadBegin)
			intervalEnd := currentTime.Sub(downloadBegin)
			if (intervalEnd - intervalStart) > minInterval {
				results = append(results, Result{Bytes: intervalBytes, IntervalStart: intervalStart, IntervalEnd: intervalEnd, Total: false})
			}
			lastCalculated = currentTime
			totalBytes += intervalBytes
			intervalBytes = 0
		}
	}

	// get last segment
	intervalStart := lastCalculated.Sub(downloadBegin)
	intervalEnd := currentTime.Sub(downloadBegin)
	if (intervalEnd - intervalStart) > minInterval {
		results = append(results, Result{Bytes: intervalBytes, IntervalStart: intervalStart, IntervalEnd: intervalEnd, Total: false})
	}

	// get total
	totalBytes += intervalBytes
	intervalEnd = currentTime.Sub(downloadBegin)
	if intervalEnd > minInterval {
		results = append(results, Result{Bytes: totalBytes, IntervalStart: 0, IntervalEnd: intervalEnd, Total: true})
	}

	return results, nil

}
