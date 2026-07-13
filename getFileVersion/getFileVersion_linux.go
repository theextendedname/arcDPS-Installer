package getfileversion

import (
	"debug/pe"
	"encoding/binary"
	"fmt"
)

const (
	resourceDirectoryIndex = 2
	versionResourceType    = 16
	resourceSubdirectory   = uint32(0x80000000)
	fixedFileInfoSignature = uint32(0xFEEF04BD)
)

func GetFileVersion(filePath string) (string, error) {
	file, err := pe.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open PE file: %w", err)
	}
	defer file.Close()

	resourceRVA, err := getResourceRVA(file)
	if err != nil {
		return "", err
	}

	resourceSection, err := sectionForRVA(file, resourceRVA)
	if err != nil {
		return "", err
	}
	sectionData, err := resourceSection.Data()
	if err != nil {
		return "", fmt.Errorf("failed to read PE resource section: %w", err)
	}

	resourceBase := int(resourceRVA - resourceSection.VirtualAddress)
	dataRVA, dataSize, err := findVersionResource(sectionData, resourceBase)
	if err != nil {
		return "", err
	}

	dataSection, err := sectionForRVA(file, dataRVA)
	if err != nil {
		return "", err
	}
	data, err := dataSection.Data()
	if err != nil {
		return "", fmt.Errorf("failed to read PE version data: %w", err)
	}
	dataOffset := int(dataRVA - dataSection.VirtualAddress)
	if dataOffset < 0 || dataOffset > len(data) || uint64(dataOffset)+uint64(dataSize) > uint64(len(data)) {
		return "", fmt.Errorf("invalid PE version resource bounds")
	}

	return parseFixedFileVersion(data[dataOffset : dataOffset+int(dataSize)])
}

func getResourceRVA(file *pe.File) (uint32, error) {
	switch header := file.OptionalHeader.(type) {
	case *pe.OptionalHeader32:
		return header.DataDirectory[resourceDirectoryIndex].VirtualAddress, nil
	case *pe.OptionalHeader64:
		return header.DataDirectory[resourceDirectoryIndex].VirtualAddress, nil
	default:
		return 0, fmt.Errorf("unsupported PE optional header")
	}
}

func sectionForRVA(file *pe.File, rva uint32) (*pe.Section, error) {
	for _, section := range file.Sections {
		size := section.VirtualSize
		if section.Size > size {
			size = section.Size
		}
		if rva >= section.VirtualAddress && rva-section.VirtualAddress < size {
			return section, nil
		}
	}
	return nil, fmt.Errorf("PE resource address 0x%x is outside all sections", rva)
}

func findVersionResource(data []byte, base int) (uint32, uint32, error) {
	typeDirectory, err := findDirectoryEntry(data, base, base, versionResourceType)
	if err != nil {
		return 0, 0, err
	}
	languageDirectory, err := firstDirectoryEntry(data, base, typeDirectory)
	if err != nil {
		return 0, 0, err
	}
	dataEntry, err := firstDataEntry(data, base, languageDirectory)
	if err != nil {
		return 0, 0, err
	}
	if dataEntry < 0 || dataEntry+16 > len(data) {
		return 0, 0, fmt.Errorf("invalid PE version data entry")
	}
	return binary.LittleEndian.Uint32(data[dataEntry : dataEntry+4]),
		binary.LittleEndian.Uint32(data[dataEntry+4 : dataEntry+8]), nil
}

func findDirectoryEntry(data []byte, base, directory, id int) (int, error) {
	entries, err := directoryEntries(data, directory)
	if err != nil {
		return 0, err
	}
	for _, entry := range entries {
		name := binary.LittleEndian.Uint32(data[entry : entry+4])
		offset := binary.LittleEndian.Uint32(data[entry+4 : entry+8])
		if name&resourceSubdirectory == 0 && int(name&0xffff) == id && offset&resourceSubdirectory != 0 {
			return base + int(offset&^resourceSubdirectory), nil
		}
	}
	return 0, fmt.Errorf("PE version resource not found")
}

func firstDirectoryEntry(data []byte, base, directory int) (int, error) {
	entries, err := directoryEntries(data, directory)
	if err != nil || len(entries) == 0 {
		return 0, fmt.Errorf("invalid PE version resource directory")
	}
	offset := binary.LittleEndian.Uint32(data[entries[0]+4 : entries[0]+8])
	if offset&resourceSubdirectory == 0 {
		return 0, fmt.Errorf("invalid PE version resource tree")
	}
	return base + int(offset&^resourceSubdirectory), nil
}

func firstDataEntry(data []byte, base, directory int) (int, error) {
	entries, err := directoryEntries(data, directory)
	if err != nil || len(entries) == 0 {
		return 0, fmt.Errorf("invalid PE version language directory")
	}
	offset := binary.LittleEndian.Uint32(data[entries[0]+4 : entries[0]+8])
	if offset&resourceSubdirectory != 0 {
		return 0, fmt.Errorf("invalid PE version data entry")
	}
	return base + int(offset), nil
}

func directoryEntries(data []byte, directory int) ([]int, error) {
	if directory < 0 || directory+16 > len(data) {
		return nil, fmt.Errorf("invalid PE resource directory")
	}
	count := int(binary.LittleEndian.Uint16(data[directory+12:directory+14])) +
		int(binary.LittleEndian.Uint16(data[directory+14:directory+16]))
	end := directory + 16 + count*8
	if end > len(data) {
		return nil, fmt.Errorf("invalid PE resource directory entries")
	}
	entries := make([]int, count)
	for i := range entries {
		entries[i] = directory + 16 + i*8
	}
	return entries, nil
}

func parseFixedFileVersion(data []byte) (string, error) {
	if len(data) < 6 {
		return "", fmt.Errorf("invalid PE version information")
	}

	offset := 6
	for {
		if offset+2 > len(data) {
			return "", fmt.Errorf("invalid PE version information key")
		}
		character := binary.LittleEndian.Uint16(data[offset : offset+2])
		offset += 2
		if character == 0 {
			break
		}
	}
	offset = (offset + 3) &^ 3
	if offset+16 > len(data) || binary.LittleEndian.Uint32(data[offset:offset+4]) != fixedFileInfoSignature {
		return "", fmt.Errorf("invalid PE fixed file information")
	}

	versionMS := binary.LittleEndian.Uint32(data[offset+8 : offset+12])
	versionLS := binary.LittleEndian.Uint32(data[offset+12 : offset+16])
	return fmt.Sprintf("%d.%d.%d", versionMS>>16, versionMS&0xffff, versionLS>>16), nil
}
