// Paquete principal de la aplicación
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec" // Nuevo: Para ejecutar el archivo .bat
	"path/filepath"
	"strings" // Nuevo: Para manipulación de strings

	// Importaciones de Fyne para la GUI
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const VERSION = "WinPass v3.1"

// Constante que define el contenido de la plantilla del archivo .bat
// Usamos %%~dp0 para escapar el % en Go y que el archivo final contenga %~dp0
const batContent = `@echo off
Set __COMPAT_LAYER=RunAsInvoker
Start "" "%%~dp0%s"`

// Variable de estado para la ruta final a ejecutar.
var pathForExecution string = ""

// showMessage muestra un cuadro de diálogo simple para notificaciones
func showMessage(w fyne.Window, title string, message string) {
	d := dialog.NewInformation(title, message, w)
	d.Show()
}

// createBatFile es la lógica principal para generar el archivo .bat
func createBatFile(w fyne.Window, exePath string) (bool, string) {
	// 1. Verificar si la ruta del .exe es válida
	if exePath == "" || filepath.Ext(exePath) != ".exe" {
		showMessage(w, "ERROR", "Please, select a valid binary (.exe files only)")
		return false, ""
	}

	dirPath := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	// 2. Definir la ruta del nuevo archivo .bat
	baseName := exeName[:len(exeName)-len(filepath.Ext(exeName))]
	batFileName := fmt.Sprintf("%_Payload.bat", baseName)
	batPath := filepath.Join(dirPath, batFileName)

	// 3. Formatear el contenido
	content := fmt.Sprintf(batContent, exeName)

	// 4. Escribir el contenido en el nuevo archivo .bat
	err := ioutil.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		showMessage(w, "ERROR", fmt.Sprintf("Payload couldn't be created%v", err))
		return false, ""
	}

	// 5. Éxito
	return true, batPath
}

// createGeneratorTab construye el contenido de la pestaña "Generar BAT"
func createGeneratorTab(w fyne.Window, pathEntry *widget.Entry, executeButton *widget.Button) fyne.CanvasObject {
	var selectedExePath string // Ruta del .exe seleccionado en esta pestaña

	// Etiqueta dinámica para mostrar la ruta de guardado
	savePathLabel := widget.NewLabel("Saving Path: No Selected")

	// Botón principal para generar el archivo .bat
	generateButton := widget.NewButton("Generate Payload", func() {
		success, path := createBatFile(w, selectedExePath)
		if success {
			pathForExecution = path
			pathEntry.SetText(path) // Muestra la ruta generada en el campo principal
			executeButton.Enable()  // Habilita el botón de ejecución
			showMessage(w, "Sucess", fmt.Sprintf("Payload sucessfully created, now go to the run tab to run it"))
		}
	})

	// Botón para seleccionar el archivo .exe
	selectExeButton := widget.NewButtonWithIcon("Select .exe file", theme.FileIcon(), func() {
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				return
			}
			selectedExePath = read.URI().Path()

			// Calcular la ruta del .bat esperado para la pre-ejecución
			basePath := strings.TrimSuffix(selectedExePath, filepath.Ext(selectedExePath))
			expectedBatPath := basePath + "_Payload.bat"

			// Actualiza la etiqueta de ruta de guardado para el usuario
			savePathLabel.SetText(fmt.Sprintf("Saving Path:\n%s", filepath.FromSlash(expectedBatPath)))

			// Si el .bat esperado existe, actualizamos el estado de ejecución
			if _, err := os.Stat(filepath.FromSlash(expectedBatPath)); err == nil {
				pathForExecution = expectedBatPath
				pathEntry.SetText(expectedBatPath)
				executeButton.Enable()
			} else {
				// Si no existe, preparamos la ejecución para la ruta del .exe (para la generación)
				pathForExecution = ""
				pathEntry.SetText(selectedExePath)
				executeButton.Disable()
			}

		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
		fd.Show()
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("WinPass | Paylaod Generation", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Step 1: Select .exe file"),
		selectExeButton,
		widget.NewSeparator(),
		widget.NewLabel("Step 2: Generate payload in the same folder"),
		generateButton,
		widget.NewSeparator(),
		savePathLabel, // Muestra la ruta de guardado dinámica
	)
}

// createExecutionTab construye el contenido de la pestaña "Ejecutar BAT"
func createExecutionTab(w fyne.Window, pathEntry *widget.Entry, executeButton *widget.Button) fyne.CanvasObject {
	// Botón para seleccionar un archivo .bat manualmente
	selectBatButton := widget.NewButtonWithIcon("Manual Payload Selection", theme.DocumentIcon(), func() {
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				return
			}
			batPath := read.URI().Path()
			pathEntry.SetText(batPath)

			pathForExecution = batPath // Establece la ruta para la ejecución
			executeButton.Enable()     // Habilita la ejecución

		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".bat"}))
		fd.Show()
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("WinPass | Paylaod Run", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Step 1: Select the desrired _Payload.bat file"),
		selectBatButton,
		widget.NewSeparator(),
		widget.NewLabel("Step 2: Click the button to run it"),
		executeButton,
		widget.NewSeparator(),
		widget.NewLabel("Paylaod loaded to run: "),
		pathEntry,
	)
}

func main() {
	a := app.New()
	w := a.NewWindow(VERSION)
	w.Resize(fyne.NewSize(750, 400))
	w.SetFixedSize(true)

	// Campos de entrada y botones que se comparten entre pestañas
	pathEntry := widget.NewEntry()
	pathEntry.PlaceHolder = "Path to the paylaod file to run"
	pathEntry.Disable() // Siempre desactivado para evitar edición manual y corrupción de ruta

	// Lógica de ejecución, compartida por ambas pestañas
	executeButton := widget.NewButtonWithIcon("Run Payload", theme.MediaPlayIcon(), func() {
		if pathForExecution == "" {
			showMessage(w, "ERROR", "You need to select a valid payload to load")
			return
		}

		// ******* LÓGICA DE EJECUCIÓN (USANDO pathForExecution) *******
		finalPath := pathForExecution

		// 2. CORRECCIÓN URI A PATH NATIVO: LIMPIEZA CRÍTICA
		finalPath = filepath.FromSlash(finalPath)

		if len(finalPath) > 1 && finalPath[0] == '\\' && finalPath[2] == ':' {
			finalPath = finalPath[1:]
		}

		if len(finalPath) > 1 && finalPath[1] == ':' {
			finalPath = strings.ToUpper(finalPath[:1]) + finalPath[1:]
		}

		// ******* EJECUCIÓN MEJORADA CON EXPLORER.EXE *******
		cmd := exec.Command("explorer", finalPath)

		if err := cmd.Start(); err != nil {
			showMessage(w, "ERROR", fmt.Sprintf("Payload couldnt be launched: Error: %v", err))
			return
		}

		showMessage(w, "SUCESS", "Paylaod runned sucessfully")
		// *************************
	})
	executeButton.Disable() // Deshabilitado al inicio

	// ------------------------------------------
	// 				CONFIGURACIÓN DE PESTAÑAS
	// ------------------------------------------

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("", theme.ContentAddIcon(), // Título vacío para mostrar solo el icono
			container.NewPadded(createGeneratorTab(w, pathEntry, executeButton)),
		),
		container.NewTabItemWithIcon("", theme.MediaPlayIcon(), // Título vacío para mostrar solo el icono
			container.NewPadded(createExecutionTab(w, pathEntry, executeButton)),
		),
		// Se pueden añadir más pestañas aquí si se necesita, por ejemplo, Ayuda/Acerca de
	)

	// Hacemos que los tabs se muestren en el lateral (izquierda)
	tabs.SetTabLocation(container.TabLocationLeading)

	w.SetContent(tabs)
	w.ShowAndRun()
}
