//go:build windows

package api

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/text/encoding/unicode"
	"image"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

const (
	SHGFI_ICON      = 0x000000100
	SHGFI_SMALLICON = 0x000000001
)

var (
	modshell32 = windows.NewLazySystemDLL("shell32.dll")
	moduser32  = windows.NewLazySystemDLL("user32.dll")
	modgdi32   = windows.NewLazySystemDLL("Gdi32.dll")

	procSHGetFileInfoW = modshell32.NewProc("SHGetFileInfoW")
	procExtractIconW   = modshell32.NewProc("ExtractIconW")
	procExtractIconExW = modshell32.NewProc("ExtractIconExW")
	procGetIconInfo    = moduser32.NewProc("GetIconInfo")
	procGetObject      = modgdi32.NewProc("GetObjectW")
	procGetBitmapBits  = modgdi32.NewProc("GetBitmapBits")
)

const (
	MAX_PATH = 260
)

type HICON syscall.Handle
type HBITMAP syscall.Handle
type DWORD uint32
type WCHAR uint16
type LONG int32
type WORD uint16

type SHFILEINFOW struct {
	hIcon         HICON
	iIcon         int32
	dwAttributes  DWORD
	szDisplayName [MAX_PATH]WCHAR
	szTypeName    [80]WCHAR
}

type ICONINFO struct {
	fIcon    bool
	xHotspot DWORD
	yHotspot DWORD
	hbmMask  HBITMAP
	hbmColor HBITMAP
}

type BITMAP struct {
	bmType       LONG
	bmWidth      LONG
	bmHeight     LONG
	bmWidthBytes LONG
	bmPlanes     WORD
	bmBitsPixel  WORD
	bmBits       unsafe.Pointer
}

func shGetFileInfo(pszPath *uint16, dwFileAttributes DWORD, psfi *SHFILEINFOW, cbFileInfo uintptr, uFlags uint) (err error) {
	r1, _, e1 := procSHGetFileInfoW.Call(
		uintptr(unsafe.Pointer(pszPath)),
		uintptr(dwFileAttributes),
		uintptr(unsafe.Pointer(psfi)),
		cbFileInfo,
		uintptr(uFlags),
	)

	if r1 == 0 {
		err = e1
	}
	return
}

func GetIconInfo(pszPath HICON, piconinfo *ICONINFO) (err error) {
	r1, _, e1 := procGetIconInfo.Call(uintptr(pszPath), uintptr(unsafe.Pointer(piconinfo)))

	if r1 == 0 {
		err = e1
	}
	return
}

func GetObject(h HBITMAP, c uintptr, pv *BITMAP) (err error) {
	r1, _, e1 := procGetObject.Call(uintptr(h), uintptr(c), uintptr(unsafe.Pointer(pv)))

	if r1 == 0 {
		err = e1
	}
	return
}

func GetBitmapBits(hbit HBITMAP, cb LONG, lpvBits []byte) (err error) {
	var _p0 *byte
	if len(lpvBits) > 0 {
		_p0 = &lpvBits[0]
	}
	r1, _, e1 := procGetBitmapBits.Call(uintptr(hbit), uintptr(cb), uintptr(unsafe.Pointer(_p0)))

	if r1 == 0 {
		err = e1
	}
	return
}

func ExtractIconEx(pszExeFileName *uint16, nIconIndex int32, iconLarge *HICON, iconSmall *HICON, nIcons uint32) (err error) {
	r1, _, e1 := procExtractIconExW.Call(
		uintptr(unsafe.Pointer(pszExeFileName)),
		uintptr(nIconIndex),
		uintptr(unsafe.Pointer(iconLarge)),
		uintptr(unsafe.Pointer(iconSmall)),
		uintptr(nIcons))

	if r1 == 0 {
		err = e1
	}

	return
}

// ShortcutIconTarget
// Reference: https://github.com/libyal/liblnk/blob/main/documentation/Windows%20Shortcut%20File%20(LNK)%20format.asciidoc#file_attribute_flags
func ShortcutIconTarget(shortcut string) (retIconPath string, retIconIndex int32, err error) {
	file, err := os.OpenFile(shortcut, os.O_RDONLY, 0)
	if err != nil {
		return
	}

	// Assert that this is a shortcut file
	var headerSize uint32
	binary.Read(file, binary.LittleEndian, &headerSize)
	if headerSize != 0x4c {
		err = fmt.Errorf("the header doesn't match a windows shortcut file")
		return
	}

	// Skip to the data flags
	if _, err = file.Seek(16, 1); err != nil {
		return
	}

	var flags int32
	if err = binary.Read(file, binary.LittleEndian, &flags); err != nil {
		return
	}

	hasTargetIDList := flags&0x01 != 0
	hasLinkInfo := flags&0x02 != 0
	hasName := flags&0x04 != 0
	hasRelativePath := flags&0x08 != 0
	hasWorkingDir := flags&0x10 != 0
	hasCommandLineArgs := flags&0x20 != 0
	hasIconLocation := flags&0x40 != 0
	isUnicode := flags&0x80 != 0
	hasExpIcon := flags&0x00004000 != 0

	// to the end of the header
	if _, err = file.Seek(56, 0); err != nil {
		return
	}

	if err = binary.Read(file, binary.LittleEndian, &retIconIndex); err != nil {
		return
	}

	_, err = file.Seek(76, 0)

	if hasTargetIDList {
		// consume the target id list
		var size16 uint16
		binary.Read(file, binary.LittleEndian, &size16)
		_, err = file.Seek(int64(size16), 1)
		if err != nil {
			err = fmt.Errorf("unable to skip shell item identifiers list")
			return
		}
	}

	if hasLinkInfo {
		var headerSize, locationInfoSize, flags, volumeOffset, localPathOffset, networkShareOffset, commonPathOffset uint32
		binary.Read(file, binary.LittleEndian, &headerSize)
		binary.Read(file, binary.LittleEndian, &locationInfoSize)
		binary.Read(file, binary.LittleEndian, &flags)
		binary.Read(file, binary.LittleEndian, &volumeOffset)
		binary.Read(file, binary.LittleEndian, &localPathOffset)
		binary.Read(file, binary.LittleEndian, &networkShareOffset)
		binary.Read(file, binary.LittleEndian, &commonPathOffset)

		if locationInfoSize > 28 {
			file.Seek(4, 1) // Skip unicode local path
		}
		if locationInfoSize > 32 {
			file.Seek(4, 1) // Skip unicode local path
			// Offset to the Unicode common path
		}

		// Read location information data
		readVolumeInformation(file)

		localPathEnd := networkShareOffset
		if localPathEnd == 0 {
			localPathEnd = commonPathOffset
		}
		localPath := make([]byte, localPathEnd-localPathOffset)
		file.Read(localPath)

		// Read network share information
		if networkShareOffset > 0 {
			readNetworkShareInfo(file)
		}

		commonPath := make([]byte, headerSize-commonPathOffset)
		file.Read(commonPath)
		// Todo: is the common path actually used?

		if !hasIconLocation && !hasExpIcon {
			// return the target location
			retIconPath = string(localPath[0 : len(localPath)-1])
			return
		}
	}

	// Read data strings
	if hasName {
		skipDataString(file, isUnicode)
	}

	if hasRelativePath {
		skipDataString(file, isUnicode)
	}

	if hasWorkingDir {
		skipDataString(file, isUnicode)
	}

	if hasCommandLineArgs {
		skipDataString(file, isUnicode)
	}

	if hasIconLocation {
		retIconPath = nextDataString(file, isUnicode)
		retIconPath, _ = registry.ExpandString(retIconPath)
		return
	}

	// Read extra data blocks until icon block
	for {
		size, signature, atEnd, _err := readExtraDataBlockHeader(file)

		if _err != nil {
			err = _err
			return
		}

		if atEnd {
			break
		}

		switch signature {
		case 0xa0000009:
			// Skip metadata
			file.Seek(int64(size), 1)
		default:
			log.Printf("Siganture: %X\n", signature)
			file.Seek(int64(size), 1)
		}
	}

	err = fmt.Errorf("could not find icon")
	return
}

func readExtraDataBlockHeader(file *os.File) (size uint32, signature uint32, atEnd bool, err error) {
	if err = binary.Read(file, binary.LittleEndian, &size); err != nil {
		return
	}

	if size == 0 {
		atEnd = true
		return
	}

	if err = binary.Read(file, binary.LittleEndian, &signature); err != nil {
		return
	}

	// compensate for size and signature
	size -= 8
	return
}

func skipDataString(file *os.File, isUnicode bool) {
	var l uint16
	binary.Read(file, binary.LittleEndian, &l)

	if isUnicode {
		l *= 2
	}

	file.Seek(int64(l), 1)
}

func nextDataString(file *os.File, isUnicode bool) string {
	var l uint16
	binary.Read(file, binary.LittleEndian, &l)
	if isUnicode {
		buff := make([]byte, l*2)
		file.Read(buff)
		win16be := unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
		result, _ := win16be.NewDecoder().Bytes(buff)
		return string(result)
	} else {
		buff := make([]byte, l)
		file.Read(buff)
		return string(buff[0 : l-1])
	}
}

func readNetworkShareInfo(file *os.File) {
	panic("Not implemented yet")
}

func readVolumeInformation(file *os.File) error {
	var size, driveType, offset, uoffset uint32
	binary.Read(file, binary.LittleEndian, &size)
	binary.Read(file, binary.LittleEndian, &driveType)
	file.Seek(4, 1) // Skip drive serial
	binary.Read(file, binary.LittleEndian, &offset)

	if offset > 16 {
		binary.Read(file, binary.LittleEndian, &uoffset)
	}

	volumeLabel := make([]byte, size-offset)

	_, err := file.Read(volumeLabel)
	if err != nil {
		return err
	}

	return nil
}

func GetIconFromResource(filename string, index int32) (i image.Image, err error) {
	var iconSmall HICON
	if err = ExtractIconEx(windows.StringToUTF16Ptr(filename), index, nil, &iconSmall, 1); err != nil {
		return
	}

	i, err = IconToImage(iconSmall)
	return
}

func GetShellIconForFile(filename string) (i image.Image, err error) {
	info := SHFILEINFOW{}

	// Fill the file info structure
	if err = shGetFileInfo(windows.StringToUTF16Ptr(filename), 0, &info, unsafe.Sizeof(info), SHGFI_ICON|SHGFI_SMALLICON); err != nil {
		return
	}

	i, err = IconToImage(info.hIcon)
	return
}

func IconToImage(iconHandle HICON) (i image.Image, err error) {
	iconInfo := ICONINFO{}
	bmp := BITMAP{}

	if err = GetIconInfo(iconHandle, &iconInfo); err != nil {
		return
	}

	if err = GetObject(iconInfo.hbmMask, unsafe.Sizeof(bmp), &bmp); err != nil {
		return
	}
	bitSize := (int(bmp.bmWidth) * int(bmp.bmHeight)) * int(bmp.bmBitsPixel) / 8
	bitBuffer := make([]byte, bitSize)
	if err = GetBitmapBits(iconInfo.hbmMask, LONG(bitSize), bitBuffer); err != nil {
		return
	}

	if err = GetObject(iconInfo.hbmColor, unsafe.Sizeof(bmp), &bmp); err != nil {
		return
	}
	imageSize := (int(bmp.bmWidth) * int(bmp.bmHeight)) * int(bmp.bmBitsPixel) / 8
	byteBuffer := make([]byte, imageSize)
	if err = GetBitmapBits(iconInfo.hbmColor, LONG(imageSize), byteBuffer); err != nil {
		return
	}

	// Check that there is an alpha channel set for any pixels in the bitmap
	//hasAlpha := 0
	//for i := 0; i < imageSize; i += 4 {
	//	if byteBuffer[i+3] != 0 {
	//		hasAlpha = 1
	//		break
	//	}
	//}

	for i := 0; i < imageSize; i += 4 {
		// Convert from windows BGRA -> Golang RGBA
		b := byteBuffer[i]
		byteBuffer[i] = byteBuffer[i+2]
		byteBuffer[i+2] = b
	}

	i = &image.NRGBA{
		Pix:    byteBuffer,
		Stride: int(bmp.bmWidth) * 4,
		Rect:   image.Rectangle{image.Point{0, 0}, image.Point{int(bmp.bmWidth), int(bmp.bmHeight)}},
	}

	return
}

// Public API functions

// GetShellIconImage is the public API of the OS package
func GetShellIconImage(path string) (im image.Image, err error) {
	var target string
	var index int32

	ext := filepath.Ext(path)

	if ext == ".lnk" {
		target, index, err = ShortcutIconTarget(path)
		if err != nil {
			return
		}

		ext := filepath.Ext(target)
		if ext == ".dll" || ext == ".cpl" {
			// fix weird indexes
			if index == -1 {
				index = 0
			}
			im, err = GetIconFromResource(target, index)
			return
		}
	} else {
		target = path
	}

	im, err = GetShellIconForFile(target)
	return
}

func ShellExecuteItem(v string) {
	var program16 *uint16
	var cmd16 *uint16
	var args16 *uint16
	var cwd16 *uint16

	program16, _ = windows.UTF16PtrFromString(v)
	cmd16, _ = windows.UTF16PtrFromString("open")

	go func() {
		err := windows.ShellExecute(0, cmd16, program16, args16, cwd16, windows.SW_SHOWNORMAL)
		if err != nil {
			log.Fatal(err)
		}
	}()
}
