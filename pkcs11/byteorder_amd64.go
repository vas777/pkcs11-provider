//
// byteorder_amd64.go
//
// Copyright (c) 2021 Markku Rossi
//
// All rights reserved.
//

package pkcs11

import (
	"encoding/binary"
)

var hbo = binary.LittleEndian