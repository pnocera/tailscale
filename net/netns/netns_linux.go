// Copyright (c) 2020 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build linux,!android

package netns

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
	"tailscale.com/hostinfo"
	"tailscale.com/net/interfaces"
)

// tailscaleBypassMark is the mark indicating that packets originating
// from a socket should bypass Tailscale-managed routes during routing
// table lookups.
//
// Keep this in sync with tailscaleBypassMark in
// wgengine/router/router_linux.go.
const tailscaleBypassMark = 0x80000

// socketMarkWorksOnce is the sync.Once & cached value for useSocketMark.
var socketMarkWorksOnce struct {
	sync.Once
	v bool
}

// socketMarkWorks returns whether SO_MARK works. If unsure, it returns true as the
// consequences of returning an incorrect false are worse.
func socketMarkWorks() bool {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:1")
	if err != nil {
		return true
	}

	sConn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return true
	}
	defer sConn.Close()

	rConn, err := sConn.SyscallConn()
	if err != nil {
		return true
	}

	var sockErr error
	err = rConn.Control(func(fd uintptr) {
		sockErr = setBypassMark(fd)
	})
	if err != nil || sockErr != nil {
		return false
	}

	return true
}

// useSocketMark reports whether SO_MARK works.
// If it doesn't, we have to use SO_BINDTODEVICE on our sockets instead.
func useSocketMark() bool {
	socketMarkWorksOnce.Do(func() {
		ipRuleWorks := exec.Command("ip", "rule").Run() == nil
		socketMarkWorksOnce.v = ipRuleWorks && socketMarkWorks()
	})
	return socketMarkWorksOnce.v
}

// ignoreErrors returns true if we should ignore setsocketopt errors in
// this instance.
func ignoreErrors() bool {
	// If we're in a test, ignore errors. Assume the test knows
	// what it's doing and will do its own skips or permission
	// checks if it's setting up a world that needs netns to work.
	// But by default, assume that tests don't need netns and it's
	// harmless to ignore the sockopts failing.
	if hostinfo.GetEnvType() == hostinfo.TestCase {
		return true
	}
	if os.Getuid() != 0 {
		// only root can manipulate these socket flags
		return true
	}
	return false
}

// control marks c as necessary to dial in a separate network namespace.
//
// It's intentionally the same signature as net.Dialer.Control
// and net.ListenConfig.Control.
func control(network, address string, c syscall.RawConn) error {
	var sockErr error
	err := c.Control(func(fd uintptr) {
		// if the user specified the IFC they want, use SO_BINDTODEVICE
		specificInterface := (os.Getenv("TS_NETNS_BIND_IFC") != "")

		if hostinfo.GetEnvType() == hostinfo.TestCase {
			sockErr = os.ErrNotExist
		} else if useSocketMark() && !specificInterface {
			sockErr = setBypassMark(fd)
		} else {
			sockErr = bindToDevice(fd)
		}
	})
	if err != nil {
		return fmt.Errorf("RawConn.Control on %T: %w", c, err)
	}
	if sockErr != nil && ignoreErrors() {
		// TODO(bradfitz): maybe log once? probably too spammy for e.g. CLI tools like tailscale netcheck.
		return nil
	}
	return sockErr
}

func setBypassMark(fd uintptr) error {
	if err := unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_MARK, tailscaleBypassMark); err != nil {
		return fmt.Errorf("setting SO_MARK bypass: %w", err)
	}
	return nil
}

func bindToDevice(fd uintptr) error {
	ifc, err := interfaces.DefaultRouteInterface()
	if err != nil {
		// Make sure we bind to *some* interface,
		// or we could get a routing loop.
		// "lo" is always wrong, but if we don't have
		// a default route anyway, it doesn't matter.
		ifc = "lo"
	}
	bind_specific_ifc := os.Getenv("TS_NETNS_BIND_IFC")
	if bind_specific_ifc != "" {
		log.Printf("netns: TS_NETNS_BIND_IFC=%s", bind_specific_ifc)
		ifc = bind_specific_ifc
	}
	if err := unix.SetsockoptString(int(fd), unix.SOL_SOCKET, unix.SO_BINDTODEVICE, ifc); err != nil {
		return fmt.Errorf("setting SO_BINDTODEVICE: %w", err)
	}
	return nil
}
