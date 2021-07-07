// +build amd64 386 arm arm64 mipsle mips64le ppc64le riscv64 wasm

package netenc

import (
	"math/bits"
)

// Ntohs converts network into native/host order.
func Ntohs(v uint16) uint16 { return bits.ReverseBytes16(v) }

// Htonl converts native/host uint32 order into network order.
func Htonl(v uint32) uint32 { return bits.ReverseBytes32(v) }

// Htons converts native/host uint16 order into network order.
func Htons(v uint16) uint16 { return bits.ReverseBytes16(v) }
