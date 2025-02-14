package folderpicker
//https://github.com/oliverpool/go-folderpicker/blob/master/folderpicker_windows.go
import (
	"errors"
	"syscall"
	"github.com/lxn/win"
)

type winFolderPicker struct {
	message *uint16
}

type OleError struct {
	error
	Result win.HRESULT
}
type SHGetPathFromIDListError struct {
	error
}

func utf16ToString(s []uint16) string {
	return syscall.UTF16ToString(s)
}

func utf16PtrFromString(s string) (*uint16, error) {
	return syscall.UTF16PtrFromString(s)
}

func NewOleError(msg string, hr win.HRESULT) OleError {
	return OleError{
		errors.New(msg),
		hr,
	}
}

func (winFolderPicker) OleInitialize() error {
	// Calling OleInitialize (or similar) is required for BIF_NEWDIALOGSTYLE.
	if hr := win.OleInitialize(); hr != win.S_OK && hr != win.S_FALSE {
		return NewOleError("OleInitialize failed", hr)
	}
	return nil
}

func (winFolderPicker) OleUninitialize() {
	win.OleUninitialize()
}

// ExpandPath expands the pidl (identifier list) path  to a file system path.
// See https://msdn.microsoft.com/en-us/library/windows/desktop/bb762194(v=vs.85).aspx
func (winFolderPicker) ExpandPath(pidl uintptr) (string, error) {
	var path [win.MAX_PATH]uint16
	if !win.SHGetPathFromIDList(pidl, &path[0]) {
		return "", SHGetPathFromIDListError{errors.New("SHGetPathFromIDList failed")}
	}
	return utf16ToString(path[:]), nil
}

const (
	// https://msdn.microsoft.com/en-us/library/aa452874.aspx
	BFFM_SELCHANGED = 2
	// https://msdn.microsoft.com/en-us/library/aa452872.aspx
	BFFM_ENABLEOK = win.WM_USER + 101

	// https://msdn.microsoft.com/en-us/library/windows/desktop/bb773205(v=vs.85).aspx
	BIF_NEWDIALOGSTYLE = 0x00000040
	BIF_SHAREABLE      = 0x00008000
)

func (wfp winFolderPicker) DisableSelectionOnInvalidPath(hwnd win.HWND, msg uint32, lp, wp uintptr) uintptr {
	if msg == BFFM_SELCHANGED {
		var enabled uintptr
		if _, err := wfp.ExpandPath(lp); err == nil {
			enabled = 1
		}
		win.SendMessage(hwnd, BFFM_ENABLEOK, 0, enabled)
	}
	return 0
}

func (wfp *winFolderPicker) setMessage(msg string) error {
	ptr, err := utf16PtrFromString(msg)
	wfp.message = ptr
	return err
}

func (wfp winFolderPicker) Prompt(msg string) (folder string, err error) {
	if err = wfp.setMessage(msg); err != nil {
		return
	}

	if err = wfp.OleInitialize(); err != nil {
		return
	}
	defer wfp.OleUninitialize()

	pidl := win.SHBrowseForFolder(&win.BROWSEINFO{
		LpszTitle: wfp.message,
		UlFlags:   BIF_NEWDIALOGSTYLE + BIF_SHAREABLE,
		Lpfn:      syscall.NewCallback(wfp.DisableSelectionOnInvalidPath),
	})

	if pidl == 0 {
		return
	}
	defer win.CoTaskMemFree(pidl)

	return wfp.ExpandPath(pidl)
}

func pickFolder(msg string) (folder string, err error) {
	wfp := winFolderPicker{}
	return wfp.Prompt(msg)
}
