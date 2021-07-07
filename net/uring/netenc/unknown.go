// +build !mips,!mips64,!ppc64,!s390x,!amd64,!386,!arm,!arm64,!mipsle,!mips64le,!ppc64le,!riscv64,!wasm

package netenc

import "C"

// Ntohs converts network into native/host order.
func Ntohs(v uint16) uint16 { return uint16(C.ntohs(C.uint16_t(v))) }

// Htonl converts native/host uint32 order into network order.
func Htonl(v uint32) uint32 { return uint32(C.htonl(c.uint32_t(v))) }

// Htons converts native/host uint16 order into network order.
func Htons(v uint16) uint16 { return uint16(C.htons(c.uint16_t(v))) }
