// umsgpack is a maximally simple implementation of MessagePack.
// It deals only in basic types (e.g., maps, arrays) but supports extensions.
package umsgpack

import (
	"strconv"
)

func init() {
	if strconv.IntSize < 64 {
		panic("umsgpack requires at least 64-bit int!")
	}
}
