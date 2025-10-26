package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func showMessage(title, message string) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	msgPtr, _ := syscall.UTF16PtrFromString(message)
	messageBox.Call(0, uintptr(unsafe.Pointer(msgPtr)), uintptr(unsafe.Pointer(titlePtr)), 0)
}

func createBatFile(exePath string) (bool, string) {
	if exePath == "" || filepath.Ext(exePath) != ".exe" {
		return false, ""
	}
	dir := filepath.Dir(exePath)
	name := filepath.Base(exePath)
	batName := fmt.Sprintf("%s_RunAsInvoker.bat", strings.TrimSuffix(name, ".exe"))
	batPath := filepath.Join(dir, batName)
	content := fmt.Sprintf("@echo off\nSet __COMPAT_LAYER=RunAsInvoker\nStart \"\" \"%%~dp0%s\"", name)
	err := os.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		return false, ""
	}
	return true, batPath
}

func openFileDialog() string {
	var ofn struct {
		lStructSize uint32
		hwndOwner   uintptr
		hInstance   uintptr
		lpstrFilter *uint16
		lpstrFile   *uint16
		nMaxFile    uint32
		// ... other fields omitted for brevity
	}
	fileBuf := make([]uint16, 260)
	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.lpstrFile = &fileBuf[0]
	ofn.nMaxFile = 260

	comdlg32 := windows.NewLazySystemDLL("comdlg32.dll")
	getOpenFileName := comdlg32.NewProc("GetOpenFileNameW")
	ret, _, _ := getOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(fileBuf)
}

func main() {
	selectedExePath := openFileDialog()
	if selectedExePath == "" {
		showMessage("Sin selección", "No se seleccionó ningún archivo .exe.")
		return
	}

	ok, batPath := createBatFile(selectedExePath)
	if !ok {
		showMessage("Error", "No se pudo generar el archivo .bat.")
		return
	}

	err := exec.Command("cmd", "/c", batPath).Start()
	if err != nil {
		showMessage("Error al ejecutar", fmt.Sprintf("No se pudo ejecutar el archivo .bat:\n%v", err))
		return
	}

	showMessage("Éxito", "Archivo .bat generado y ejecutado correctamente.")
}
