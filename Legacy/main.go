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

// Estructura OPENFILENAME completa para compatibilidad con Windows (incluyendo campos para WINVER >= 0x0400)
type OPENFILENAME struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	FlagsEx           uint32
}

func showMessage(title, message string) {
	user32 := windows.NewLazySystemDLL("user32.dll")
	messageBox := user32.NewProc("MessageBoxW")
	titlePtr, _ := windows.UTF16PtrFromString(title)
	msgPtr, _ := windows.UTF16PtrFromString(message)
	messageBox.Call(0, uintptr(unsafe.Pointer(msgPtr)), uintptr(unsafe.Pointer(titlePtr)), 0)
}

func createBatFile(exePath string) (bool, string) {
	if exePath == "" || filepath.Ext(exePath) != ".exe" {
		return false, ""
	}
	dir := filepath.Dir(exePath)
	name := filepath.Base(exePath)
	baseName := strings.TrimSuffix(name, ".exe")
	batName := baseName + "_RunAsInvoker.bat"
	batPath := filepath.Join(dir, batName)
	content := fmt.Sprintf("@echo off\nSet __COMPAT_LAYER=RunAsInvoker\nStart \"\" \"%%~dp0%s\"", name)
	err := os.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		return false, ""
	}
	return true, batPath
}

func openFileDialog(filterDesc, filterPat string) string {
	filterStr := filterDesc + "\x00" + filterPat + "\x00All Files\x00*.*\x00\x00"
	filter16, err := syscall.UTF16FromString(filterStr)
	if err != nil || len(filter16) == 0 {
		return "" // Fallback si falla la conversión
	}

	var ofn OPENFILENAME
	fileBuf := make([]uint16, 260)
	titleBuf := make([]uint16, 260)

	ofn.lStructSize = uint32(unsafe.Sizeof(ofn))
	ofn.lpstrFilter = &filter16[0]
	ofn.lpstrFile = &fileBuf[0]
	ofn.nMaxFile = uint32(len(fileBuf))
	ofn.lpstrFileTitle = &titleBuf[0]
	ofn.nMaxFileTitle = uint32(len(titleBuf))
	ofn.nFilterIndex = 1
	ofn.flags = 0x00080000 | 0x00001000 | 0x00200000 // OFN_FILEMUSTEXIST | OFN_PATHMUSTEXIST | OFN_LONGNAMES
	// Inicializar campos adicionales a 0
	ofn.pvReserved = 0
	ofn.dwReserved = 0
	ofn.FlagsEx = 0

	comdlg32 := windows.NewLazySystemDLL("comdlg32.dll")
	getOpenFileName := comdlg32.NewProc("GetOpenFileNameW")
	ret, _, _ := getOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return ""
	}

	// Buscar el terminador null en fileBuf para extraer la ruta correctamente
	path := syscall.UTF16ToString(fileBuf)
	return strings.TrimRight(path, "\x00") // Limpia terminadores extras si hay
}

func openExeDialog() string {
	return openFileDialog("Executable Files", "*.exe")
}

func openBatDialog() string {
	return openFileDialog("Batch Files", "*.bat")
}

func executeBat(batPath string) bool {
	cmd := exec.Command("cmd", "/c", batPath)
	err := cmd.Start()
	if err != nil {
		showMessage("Error al ejecutar", fmt.Sprintf("No se pudo ejecutar el archivo .bat:\n%v", err))
		return false
	}
	return true
}

func main() {
	showMessage("Bienvenido", "Generador BAT (RunAsInvoker)\nSelecciona una opción en la consola.")

	for {
		fmt.Println("\n--- Menú Principal ---")
		fmt.Println("1. Generar BAT desde EXE (modo Generador)")
		fmt.Println("2. Ejecutar BAT existente (modo Ejecutar)")
		fmt.Println("3. Salir")
		fmt.Print("Elige una opción (1-3): ")

		var choice string
		fmt.Scanln(&choice)

		switch choice {
		case "1":
			fmt.Println("Modo Generador: Selecciona el archivo .exe...")
			selectedExePath := openExeDialog()
			if selectedExePath == "" {
				showMessage("Sin selección", "No se seleccionó ningún archivo .exe.")
				continue
			}

			fmt.Println("Generando BAT...")
			ok, batPath := createBatFile(selectedExePath)
			if !ok {
				showMessage("Error", "No se pudo generar el archivo .bat.")
				continue
			}

			fmt.Printf("BAT creado en: %s\n", batPath)

			// Preguntar si ejecutar
			user32 := windows.NewLazySystemDLL("user32.dll")
			messageBox := user32.NewProc("MessageBoxW")
			msgPtr, _ := windows.UTF16PtrFromString("¿Deseas ejecutar el BAT ahora?")
			titlePtr, _ := windows.UTF16PtrFromString("Ejecutar?")
			ret, _, _ := messageBox.Call(0, uintptr(unsafe.Pointer(msgPtr)), uintptr(unsafe.Pointer(titlePtr)), 4) // MB_YESNO = 4
			if ret == 6 {                                                                                          // IDYES
				if executeBat(batPath) {
					showMessage("Éxito", "Archivo .bat generado y ejecutado correctamente.")
				}
			} else {
				showMessage("Éxito", fmt.Sprintf("Archivo .bat generado correctamente:\n%s", batPath))
			}

		case "2":
			fmt.Println("Modo Ejecutar: Selecciona el archivo .bat...")
			selectedBatPath := openBatDialog()
			if selectedBatPath == "" {
				showMessage("Sin selección", "No se seleccionó ningún archivo .bat.")
				continue
			}

			fmt.Printf("Ejecutando: %s\n", selectedBatPath)
			if executeBat(selectedBatPath) {
				showMessage("Lanzamiento", "El archivo .bat se ha lanzado correctamente.")
			}

		case "3":
			showMessage("Adiós", "Gracias por usar Generador BAT.")
			return

		default:
			showMessage("Opción inválida", "Por favor, elige 1, 2 o 3.")
		}
	}
}
