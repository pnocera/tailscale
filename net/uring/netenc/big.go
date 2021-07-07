// +build mips mips64 ppc64 s390x

package netenc

// TODO how should I capitalize this

// Ntohs converts network order into native/host.
func Ntohs(v uint16) uint16 { return v }

// Htonl converts native/host uint32 order into network order.
func Htonl(v uint32) uint32 { return v }

// Htons converts native/host uint16 order into network order.
func Htons(v uint16) uint16 { return v }
