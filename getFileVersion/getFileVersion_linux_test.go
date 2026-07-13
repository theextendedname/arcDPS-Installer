package getfileversion

import (
	"encoding/binary"
	"testing"
)

func TestParseFixedFileVersion(t *testing.T) {
	data := make([]byte, 128)
	binary.LittleEndian.PutUint16(data[0:2], uint16(len(data)))
	binary.LittleEndian.PutUint16(data[2:4], 52)

	offset := 6
	for _, character := range "VS_VERSION_INFO" {
		binary.LittleEndian.PutUint16(data[offset:offset+2], uint16(character))
		offset += 2
	}
	offset += 2
	offset = (offset + 3) &^ 3

	binary.LittleEndian.PutUint32(data[offset:offset+4], fixedFileInfoSignature)
	binary.LittleEndian.PutUint32(data[offset+8:offset+12], uint32(2<<16|18))
	binary.LittleEndian.PutUint32(data[offset+12:offset+16], uint32(1<<16|7))

	version, err := parseFixedFileVersion(data)
	if err != nil {
		t.Fatalf("parseFixedFileVersion returned an error: %v", err)
	}
	if version != "2.18.1" {
		t.Fatalf("parseFixedFileVersion returned %q, want %q", version, "2.18.1")
	}
}

func TestFindVersionResource(t *testing.T) {
	data := make([]byte, 128)

	binary.LittleEndian.PutUint16(data[14:16], 1)
	binary.LittleEndian.PutUint32(data[16:20], versionResourceType)
	binary.LittleEndian.PutUint32(data[20:24], resourceSubdirectory|32)

	binary.LittleEndian.PutUint16(data[32+14:32+16], 1)
	binary.LittleEndian.PutUint32(data[32+16:32+20], 1)
	binary.LittleEndian.PutUint32(data[32+20:32+24], resourceSubdirectory|64)

	binary.LittleEndian.PutUint16(data[64+14:64+16], 1)
	binary.LittleEndian.PutUint32(data[64+16:64+20], 0x409)
	binary.LittleEndian.PutUint32(data[64+20:64+24], 96)

	binary.LittleEndian.PutUint32(data[96:100], 0x1234)
	binary.LittleEndian.PutUint32(data[100:104], 52)

	rva, size, err := findVersionResource(data, 0)
	if err != nil {
		t.Fatalf("findVersionResource returned an error: %v", err)
	}
	if rva != 0x1234 || size != 52 {
		t.Fatalf("findVersionResource returned RVA 0x%x and size %d", rva, size)
	}
}
