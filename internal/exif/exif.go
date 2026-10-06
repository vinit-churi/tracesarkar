// Package exif reads the one thing the platform needs from a photograph's own
// metadata: when it was taken.
//
// Written by hand rather than pulled in as a dependency, because the need is
// narrow — a single ASCII tag — and because every byte here arrives from the
// internet. A parser that panics on a truncated or hostile file takes the
// capture endpoint with it, and losing a photograph is the one thing this
// platform may never do.
//
// GPS is deliberately not read. Our own captures carry no GPS IFD at all: the
// client resizes to 2048 px before upload and that re-encode drops it. Reading
// a tag that is never present would be a parser nobody could test against real
// data.
package exif

import (
	"bytes"
	"encoding/binary"
	"time"
)

const (
	tagExifIFD            = 0x8769
	tagDateTimeOriginal   = 0x9003
	tagOffsetTimeOriginal = 0x9011
)

// TakenAt returns when the photograph was taken, as its own metadata records it.
//
// EXIF writes local time with no zone, so [local] is the zone it is read in.
// Mumbai is the only place this platform operates, which makes IST the right
// default — but it is passed in rather than assumed here, because a wrong zone
// turns a five-hour-old photograph into a current one.
//
// Returns false whenever the answer is not plainly there. An absent timestamp
// is an ordinary outcome, not an error: plenty of images carry no EXIF at all.
func TakenAt(image []byte, local *time.Location) (time.Time, bool) {
	tiff, order, ok := tiffBlock(image)
	if !ok {
		return time.Time{}, false
	}

	ifd0, ok := firstIFDOffset(tiff, order)
	if !ok {
		return time.Time{}, false
	}
	sub, ok := longValue(tiff, order, ifd0, tagExifIFD)
	if !ok {
		return time.Time{}, false
	}

	raw, ok := asciiValue(tiff, order, int(sub), tagDateTimeOriginal)
	if !ok {
		return time.Time{}, false
	}

	zone := local
	if offset, ok := asciiValue(tiff, order, int(sub), tagOffsetTimeOriginal); ok {
		if z, err := time.Parse("-07:00", offset); err == nil {
			zone = z.Location()
		}
	}

	t, err := time.ParseInLocation("2006:01:02 15:04:05", raw, zone)
	if err != nil {
		return time.Time{}, false
	}
	return t, true
}

// tiffBlock finds the TIFF header inside the JPEG's APP1 segment and returns
// it with its byte order.
func tiffBlock(image []byte) ([]byte, binary.ByteOrder, bool) {
	i := bytes.Index(image, []byte("Exif\x00\x00"))
	if i < 0 || i+14 > len(image) {
		return nil, nil, false
	}
	tiff := image[i+6:]
	if len(tiff) < 8 {
		return nil, nil, false
	}
	var order binary.ByteOrder
	switch {
	case tiff[0] == 'I' && tiff[1] == 'I':
		order = binary.LittleEndian
	case tiff[0] == 'M' && tiff[1] == 'M':
		order = binary.BigEndian
	default:
		return nil, nil, false
	}
	if order.Uint16(tiff[2:4]) != 42 {
		return nil, nil, false
	}
	return tiff, order, true
}

func firstIFDOffset(tiff []byte, order binary.ByteOrder) (int, bool) {
	off := int(order.Uint32(tiff[4:8]))
	if off < 8 || off+2 > len(tiff) {
		return 0, false
	}
	return off, true
}

// entry walks one directory looking for a tag, and refuses to read past the
// end of the block at every step.
func entry(tiff []byte, order binary.ByteOrder, ifd int, tag uint16) (typ uint16, count uint32, value []byte, ok bool) {
	if ifd < 0 || ifd+2 > len(tiff) {
		return 0, 0, nil, false
	}
	n := int(order.Uint16(tiff[ifd : ifd+2]))
	// A directory claiming more entries than the file can hold is corrupt.
	if n < 0 || ifd+2+n*12 > len(tiff) {
		return 0, 0, nil, false
	}
	for k := range n {
		e := ifd + 2 + k*12
		if order.Uint16(tiff[e:e+2]) != tag {
			continue
		}
		return order.Uint16(tiff[e+2 : e+4]),
			order.Uint32(tiff[e+4 : e+8]),
			tiff[e+8 : e+12], true
	}
	return 0, 0, nil, false
}

func longValue(tiff []byte, order binary.ByteOrder, ifd int, tag uint16) (uint32, bool) {
	_, _, value, ok := entry(tiff, order, ifd, tag)
	if !ok {
		return 0, false
	}
	return order.Uint32(value), true
}

func asciiValue(tiff []byte, order binary.ByteOrder, ifd int, tag uint16) (string, bool) {
	_, count, value, ok := entry(tiff, order, ifd, tag)
	if !ok || count == 0 || count > 1024 {
		return "", false
	}
	// Four bytes or fewer live in the entry itself; anything longer is an
	// offset into the block.
	raw := value
	if count > 4 {
		off := int(order.Uint32(value))
		if off < 0 || off+int(count) > len(tiff) {
			return "", false
		}
		raw = tiff[off : off+int(count)]
	}
	return string(bytes.TrimRight(raw[:min(int(count), len(raw))], "\x00 ")), true
}
