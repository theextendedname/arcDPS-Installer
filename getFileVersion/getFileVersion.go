package getfileversion

import (
        "fmt"
        "syscall"
        "unsafe"
)

var (
        modversion = syscall.NewLazyDLL("version.dll")
        procGetFileVersionInfoSize = modversion.NewProc("GetFileVersionInfoSizeW")
        procGetFileVersionInfo     = modversion.NewProc("GetFileVersionInfoW")
        procVerQueryValue          = modversion.NewProc("VerQueryValueW")
)

// VS_FIXEDFILEINFO is not directly available in the syscall package, so we define it here.
type VS_FIXEDFILEINFO struct {
        DwSignature        uint32
        DwStrucVersion     uint32
        DwFileVersionMS    uint32
        DwFileVersionLS    uint32
        DwProductVersionMS uint32
        DwProductVersionLS uint32
        DwFileFlagsMask    uint32
        DwFileFlags        uint32
        DwFileOS           uint32
        DwFileType         uint32
        DwFileSubtype      uint32
        DwFileDateMS       uint32
        DwFileDateLS       uint32
}

func GetFileVersion(filePath string) (string, error) {
        filePathPtr, err := syscall.UTF16PtrFromString(filePath)
        if err != nil {
                return "", fmt.Errorf("failed to convert file path to UTF16: %w", err)
        }

        size, _, err := procGetFileVersionInfoSize.Call(uintptr(unsafe.Pointer(filePathPtr)), 0)
        if size == 0 {
                return "", fmt.Errorf("failed to get file version info size: %w", err)
        }

        versionData := make([]byte, size)
        _, _, err = procGetFileVersionInfo.Call(uintptr(unsafe.Pointer(filePathPtr)), 0, uintptr(size), uintptr(unsafe.Pointer(&versionData[0])))
        if err != nil && err != syscall.Errno(0) {
                return "", fmt.Errorf("failed to get file version info: %w", err)
        }

        var subBlockPtr uintptr
        subBlockPtrPtr := uintptr(unsafe.Pointer(&subBlockPtr))
        var subBlockLen uintptr

        subBlock, err := syscall.UTF16PtrFromString("\\")
        if err != nil {
                return "", fmt.Errorf("failed to create utf16 string: %w", err)
        }

        _, _, err = procVerQueryValue.Call(uintptr(unsafe.Pointer(&versionData[0])), uintptr(unsafe.Pointer(subBlock)), subBlockPtrPtr, uintptr(unsafe.Pointer(&subBlockLen)))

        if err != nil && err != syscall.Errno(0) {
                return "", fmt.Errorf("failed to query version value: %w", err)
        }

        if subBlockLen == 0 {//return empty if not found
                return "", nil
        }

        fixedFileInfo := (*VS_FIXEDFILEINFO)(unsafe.Pointer(subBlockPtr))

        major := fixedFileInfo.DwFileVersionMS >> 16
        minor := fixedFileInfo.DwFileVersionMS & 0xFFFF
        build := fixedFileInfo.DwFileVersionLS >> 16
        revision := fixedFileInfo.DwFileVersionLS & 0xFFFF

        return fmt.Sprintf("%d.%d.%d.%d", major, minor, build, revision), nil
		//return fmt.Sprintf("%d.%d.%d", major, minor, build), nil
}
