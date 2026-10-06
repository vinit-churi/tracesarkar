package exif

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"
)

// jpegWith builds the smallest JPEG that carries an EXIF block, so the parser
// is tested against bytes rather than against a fixture nobody can read.
func jpegWith(t *testing.T, tags map[uint16]string) []byte {
	t.Helper()

	// TIFF body: IFD0 holds only a pointer to the Exif IFD, which holds the
	// string tags. Values longer than four bytes live after both directories.
	var tiff bytes.Buffer
	tiff.WriteString("II")                               // little-endian
	binary.Write(&tiff, binary.LittleEndian, uint16(42)) // magic
	binary.Write(&tiff, binary.LittleEndian, uint32(8))  // IFD0 at 8

	ifd0 := new(bytes.Buffer)
	binary.Write(ifd0, binary.LittleEndian, uint16(1))      // one entry
	binary.Write(ifd0, binary.LittleEndian, uint16(0x8769)) // Exif IFD pointer
	binary.Write(ifd0, binary.LittleEndian, uint16(4))      // LONG
	binary.Write(ifd0, binary.LittleEndian, uint32(1))
	exifIFDAt := uint32(8 + 2 + 12 + 4)
	binary.Write(ifd0, binary.LittleEndian, exifIFDAt)
	binary.Write(ifd0, binary.LittleEndian, uint32(0)) // no IFD1

	sub := new(bytes.Buffer)
	binary.Write(sub, binary.LittleEndian, uint16(len(tags)))
	valuesAt := exifIFDAt + uint32(2+12*len(tags)+4)
	var values bytes.Buffer
	keys := []uint16{0x9003, 0x9011}
	for _, tag := range keys {
		v, ok := tags[tag]
		if !ok {
			continue
		}
		binary.Write(sub, binary.LittleEndian, tag)
		binary.Write(sub, binary.LittleEndian, uint16(2)) // ASCII
		binary.Write(sub, binary.LittleEndian, uint32(len(v)+1))
		binary.Write(sub, binary.LittleEndian, valuesAt+uint32(values.Len()))
		values.WriteString(v)
		values.WriteByte(0)
	}
	binary.Write(sub, binary.LittleEndian, uint32(0))

	tiff.Write(ifd0.Bytes())
	tiff.Write(sub.Bytes())
	tiff.Write(values.Bytes())

	app1 := append([]byte("Exif\x00\x00"), tiff.Bytes()...)
	var out bytes.Buffer
	out.Write([]byte{0xFF, 0xD8}) // SOI
	out.Write([]byte{0xFF, 0xE1}) // APP1
	binary.Write(&out, binary.BigEndian, uint16(len(app1)+2))
	out.Write(app1)
	out.Write([]byte{0xFF, 0xD9}) // EOI
	return out.Bytes()
}

func TestTakenAt(t *testing.T) {
	ist := time.FixedZone("IST", 5*3600+1800)

	tests := []struct {
		name string
		tags map[uint16]string
		want time.Time
		ok   bool
	}{
		{
			// EXIF writes local time with no zone. Mumbai is the only place
			// this platform operates, so IST is the reading, and it has to be
			// stated rather than assumed silently.
			name: "a plain timestamp is read as local time",
			tags: map[uint16]string{0x9003: "2026:10:05 19:09:44"},
			want: time.Date(2026, 10, 5, 19, 9, 44, 0, ist),
			ok:   true,
		},
		{
			// When the camera does record an offset, it wins over the default.
			name: "an explicit offset is honoured",
			tags: map[uint16]string{0x9003: "2026:10:05 19:09:44", 0x9011: "+01:00"},
			want: time.Date(2026, 10, 5, 19, 9, 44, 0, time.FixedZone("", 3600)),
			ok:   true,
		},
		{name: "no exif block at all", tags: nil},
		{
			name: "a malformed timestamp is absent, not an error",
			tags: map[uint16]string{0x9003: "not a date"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var image []byte
			if tt.tags == nil {
				image = []byte{0xFF, 0xD8, 0xFF, 0xD9}
			} else {
				image = jpegWith(t, tt.tags)
			}

			got, ok := TakenAt(image, ist)
			if ok != tt.ok {
				t.Fatalf("ok = %v, want %v", ok, tt.ok)
			}
			if !tt.ok {
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

// Nothing here may panic on a hostile or truncated file. These bytes arrive
// from the internet, and a parser that panics on them takes the capture
// endpoint down with it — losing the photograph, which is hard rule 7.
func TestTakenAtSurvivesRubbish(t *testing.T) {
	good := jpegWith(t, map[uint16]string{0x9003: "2026:10:05 19:09:44"})

	cases := [][]byte{
		nil,
		{},
		{0xFF},
		{0xFF, 0xD8},
		[]byte("not a jpeg at all"),
		good[:len(good)/2],
		good[:12],
		append([]byte{0xFF, 0xD8, 0xFF, 0xE1, 0xFF, 0xFF}, []byte("Exif\x00\x00II")...),
	}
	for i, c := range cases {
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("case %d panicked: %v", i, r)
				}
			}()
			_, _ = TakenAt(c, time.UTC)
		}()
	}
}
