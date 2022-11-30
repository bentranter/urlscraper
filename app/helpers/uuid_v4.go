package helpers

import (
	"crypto/rand"
	"encoding/hex"
)

// UUID generates a UUID v4, to be used as request IDs.
func UUID() string {
	const dash byte = '-'
	var u [16]byte
	var buf [36]byte

	if _, err := rand.Read(u[:]); err != nil {
		panic(err)
	}

	// Set the version bit to v4
	u[6] = (u[6] & 0x0f) | 0x40

	// Set the variant bit to "don't care"
	u[8] = (u[8] & 0xbf) | 0x80

	hex.Encode(buf[0:8], u[0:4])
	buf[8] = dash
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = dash
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = dash
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = dash
	hex.Encode(buf[24:], u[10:])

	return string(buf[:])
}
