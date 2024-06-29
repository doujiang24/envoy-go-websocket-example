package main

import (
	"encoding/binary"

	"github.com/envoyproxy/envoy/contrib/golang/common/go/api"
)

type frame struct {
	buffer    []byte
	pos       int // where data start position
	frameType int
	frameLen  int64
	mask      bool
	maskKey   [4]byte
}

func (f *frame) GetData() []byte {
	return f.buffer[f.pos : f.pos+int(f.frameLen)]
}

func (f *frame) SetData(data []byte) {
	if f.frameLen >= 126 || len(data) >= 126 {
		// TODO: handle this case
		return
	}

	f.frameLen = int64(len(data))
	bitVal := f.frameLen
	if f.mask {
		bitVal |= maskBit
	}
	f.buffer[1] = byte(bitVal)

	f.buffer = append(f.buffer[0:f.pos], data...)
	if f.mask {
		maskBytes(f.maskKey, 0, f.buffer[f.pos:])
	}
}

func (f *frame) Bytes() []byte {
	api.LogDebugf("frame buffer: %v", f.buffer)
	return f.buffer
}

const (
	// Frame header byte 0 bits from Section 5.2 of RFC 6455
	finalBit = 1 << 7
	rsv1Bit  = 1 << 6
	rsv2Bit  = 1 << 5
	rsv3Bit  = 1 << 4

	// Frame header byte 1 bits from Section 5.2 of RFC 6455
	maskBit = 1 << 7
)

func readFrame(buffer []byte) ([]byte, *frame) {
	if len(buffer) < 2 {
		return buffer, nil
	}

	api.LogDebugf("buffer: %v", buffer)

	final := buffer[0]&finalBit != 0
	rsv1 := buffer[0]&rsv1Bit != 0
	rsv2 := buffer[0]&rsv2Bit != 0
	rsv3 := buffer[0]&rsv3Bit != 0
	mask := buffer[1]&maskBit != 0

	api.LogDebugf("final: %v, rsv1: %v, rsv2: %v, rsv3: %v, mask: %v", final, rsv1, rsv2, rsv3, mask)

	frameType := int(buffer[0] & 0xf)
	frameLen := int64(buffer[1] & 0x7f)
	pos := int64(2)
	switch frameLen {
	case 126:
		if len(buffer) < 4 {
			return buffer, nil
		}
		frameLen = int64(binary.BigEndian.Uint16(buffer[pos : pos+2]))
		pos += 2

	case 127:
		if len(buffer) < 10 {
			return buffer, nil
		}
		frameLen = int64(binary.BigEndian.Uint64(buffer[pos : pos+8]))
		pos += 8
	}

	var maskKey [4]byte
	if mask {
		if int64(len(buffer)) < pos+4+frameLen {
			return buffer, nil
		}
		maskKey = [4]byte(buffer[pos : pos+4])
		pos += 4

		maskBytes(maskKey, 0, buffer[pos:pos+frameLen])

	} else if int64(len(buffer)) < pos+frameLen {
		return buffer, nil
	}

	fr := frame{
		buffer:    buffer[0 : pos+frameLen],
		pos:       int(pos),
		frameType: frameType,
		frameLen:  frameLen,
		mask:      mask,
		maskKey:   maskKey,
	}
	api.LogDebugf("frameType: %d, frameLen: %d", frameType, frameLen)
	return buffer[pos+frameLen:], &fr
}
