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
		showMessage(w, "Error", "Por favor, selecciona un archivo ejecutable válido (.exe).")
		return false, ""
	}

	dirPath := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	// 2. Definir la ruta del nuevo archivo .bat
	baseName := exeName[:len(exeName)-len(filepath.Ext(exeName))]
	batFileName := fmt.Sprintf("%s_RunAsInvoker.bat", baseName)
	batPath := filepath.Join(dirPath, batFileName)

	// 3. Formatear el contenido
	content := fmt.Sprintf(batContent, exeName)

	// 4. Escribir el contenido en el nuevo archivo .bat
	err := ioutil.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		showMessage(w, "Error de Archivo", fmt.Sprintf("No se pudo crear el archivo .bat:\n%v", err))
		return false, ""
	}

	// 5. Éxito
	return true, batPath
}

// createGeneratorTab construye el contenido de la pestaña "Generar BAT"
func createGeneratorTab(w fyne.Window, pathEntry *widget.Entry, executeButton *widget.Button) fyne.CanvasObject {
	var selectedExePath string // Ruta del .exe seleccionado en esta pestaña

	// Botón principal para generar el archivo .bat
	generateButton := widget.NewButton("Generar Archivo .bat", func() {
		success, path := createBatFile(w, selectedExePath)
		if success {
			pathForExecution = path
			pathEntry.SetText(path) // Muestra la ruta generada en el campo principal
			executeButton.Enable()  // Habilita el botón de ejecución
			showMessage(w, "Éxito", fmt.Sprintf("¡Archivo .bat creado con éxito!\nAhora puedes ejecutarlo en la pestaña 'Ejecutar'."))
		}
	})

	// Botón para seleccionar el archivo .exe
	selectExeButton := widget.NewButtonWithIcon("Seleccionar Archivo .exe", theme.FileIcon(), func() {
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				return
			}
			selectedExePath = read.URI().Path()

			// Calcular la ruta del .bat esperado para la pre-ejecución
			basePath := strings.TrimSuffix(selectedExePath, filepath.Ext(selectedExePath))
			expectedBatPath := basePath + "_RunAsInvoker.bat"

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
		widget.NewLabelWithStyle("GENERADOR BAT", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Paso 1: Selecciona el archivo .exe que quieres modificar."),
		selectExeButton,
		widget.NewSeparator(),
		widget.NewLabel("Paso 2: Genera el archivo .bat en la misma carpeta."),
		generateButton,
		widget.NewSeparator(),
		widget.NewLabel("El archivo .bat se llamará: [Nombre_EXE]_RunAsInvoker.bat"),
	)
}

// createExecutionTab construye el contenido de la pestaña "Ejecutar BAT"
func createExecutionTab(w fyne.Window, pathEntry *widget.Entry, executeButton *widget.Button) fyne.CanvasObject {
	// Botón para seleccionar un archivo .bat manualmente
	selectBatButton := widget.NewButtonWithIcon("Seleccionar .bat manualmente", theme.DocumentIcon(), func() {
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
		widget.NewLabelWithStyle("EJECUTAR ARCHIVO", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Paso 1: Selecciona el archivo .bat que quieres ejecutar."),
		selectBatButton,
		widget.NewSeparator(),
		widget.NewLabel("Paso 2: Haz clic en el botón de abajo."),
		executeButton,
		widget.NewSeparator(),
		widget.NewLabel("Ruta actual a ejecutar:"),
		pathEntry,
	)
}

func main() {
	a := app.New()
	w := a.NewWindow("Generador BAT (RunAsInvoker)")
	w.Resize(fyne.NewSize(750, 400))
	w.SetFixedSize(true)

	// Campos de entrada y botones que se comparten entre pestañas
	pathEntry := widget.NewEntry()
	pathEntry.PlaceHolder = "Ruta del archivo .exe o .bat a ejecutar..."
	pathEntry.Disable() // Siempre desactivado para evitar edición manual y corrupción de ruta

	// Lógica de ejecución, compartida por ambas pestañas
	executeButton := widget.NewButtonWithIcon("Ejecutar .bat", theme.MediaPlayIcon(), func() {
		if pathForExecution == "" {
			showMessage(w, "Error", "Primero debes seleccionar un archivo .exe o un .bat para ejecutar.")
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

		// Línea de Debug: Muestra la ruta limpia que se está intentando ejecutar
		fmt.Println("Intentando ejecutar con ruta (debug):", finalPath)

		// ******* EJECUCIÓN MEJORADA CON EXPLORER.EXE *******
		cmd := exec.Command("explorer", finalPath)

		if err := cmd.Start(); err != nil {
			showMessage(w, "Error de Ejecución", fmt.Sprintf("No se pudo iniciar el archivo BAT usando explorer.exe. Error: %v", err))
			return
		}

		showMessage(w, "Lanzamiento", "El archivo .bat se ha lanzado correctamente.")
		// *************************
	})
	executeButton.Disable() // Deshabilitado al inicio

	// ------------------------------------------
	// 				CONFIGURACIÓN DE PESTAÑAS
	// ------------------------------------------

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("Generar", theme.ContentAddIcon(),
			container.NewPadded(createGeneratorTab(w, pathEntry, executeButton)),
		),
		container.NewTabItemWithIcon("Ejecutar", theme.MediaPlayIcon(),
			container.NewPadded(createExecutionTab(w, pathEntry, executeButton)),
		),
		// Se pueden añadir más pestañas aquí si se necesita, por ejemplo, Ayuda/Acerca de
	)

	// Hacemos que los tabs se muestren en el lateral (izquierda)
	tabs.SetTabLocation(container.TabLocationLeading)

	w.SetContent(tabs)
	w.ShowAndRun()
}
