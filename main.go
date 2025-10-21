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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Constante que define el contenido de la plantilla del archivo .bat
// FIX DEFINITIVO: Cambiamos %~dp0 por %%~dp0. En Go, un % literal debe ser escapado como %%
// cuando se usa con fmt.Sprintf. Esto evita la corrupción del formato.
const batContent = `@echo off
Set __COMPAT_LAYER=RunAsInvoker
Start "" "%%~dp0%s"` // Usamos %%~dp0 para que el archivo final contenga %~dp0

// Definición del icono de la aplicación.
var Icon = theme.FolderOpenIcon()

// Variable de estado para la ruta final a ejecutar, puede ser asignada
// por `createBatFile` o por la selección manual del usuario.
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
	// Elimina la extensión .exe para nombrar el .bat
	baseName := exeName[:len(exeName)-len(filepath.Ext(exeName))]
	batFileName := fmt.Sprintf("%s_RunAsInvoker.bat", baseName)
	batPath := filepath.Join(dirPath, batFileName)

	// 3. Formatear el contenido
	// Se pasa el nombre del ejecutable. El "%%~dp0" en batContent garantiza el %~dp0 final.
	content := fmt.Sprintf(batContent, exeName)

	// 4. Escribir el contenido en el nuevo archivo .bat
	err := ioutil.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		showMessage(w, "Error de Archivo", fmt.Sprintf("No se pudo crear el archivo .bat:\n%v", err))
		return false, ""
	}

	// 5. Éxito
	// Devolvemos la ruta final del BAT
	return true, batPath
}

func main() {
	a := app.New()
	w := a.NewWindow("Generador BAT (RunAsInvoker)")
	w.Resize(fyne.NewSize(750, 400))
	w.SetFixedSize(true)

	var selectedExePath string // Ruta del .exe seleccionado por el diálogo

	pathEntry := widget.NewEntry()
	pathEntry.PlaceHolder = "Ruta del archivo .exe o .bat a ejecutar..."
	pathEntry.Disable()

	// Botón para ejecutar el .bat (deshabilitado inicialmente)
	executeButton := widget.NewButtonWithIcon("Ejecutar .bat", theme.MediaPlayIcon(), func() {
		if pathForExecution == "" {
			showMessage(w, "Error", "Primero debes seleccionar un archivo .exe o un .bat para ejecutar.")
			return
		}

		// ******* LÓGICA DE EJECUCIÓN (USANDO pathForExecution) *******

		// 1. Tomar la ruta de ejecución
		finalPath := pathForExecution

		// 2. CORRECCIÓN URI A PATH NATIVO: LIMPIEZA CRÍTICA

		// Convertir separadores de Unix (/) a Windows (\).
		// Si el path de fyne es "/C:/...", esto resulta en "\C:\...".
		finalPath = filepath.FromSlash(finalPath)

		// Si la ruta comienza con una barra invertida, y luego sigue
		// la letra de la unidad (e.g., "\C:\..."), eliminar la primera barra.
		// Esto corrige la conversión de URI a path nativo de Fyne.
		if len(finalPath) > 1 && finalPath[0] == '\\' && finalPath[2] == ':' {
			finalPath = finalPath[1:] // Ahora: "C:\Users..."
		}

		// Aseguramos que la letra de la unidad esté en mayúsculas (C: en lugar de c:)
		if len(finalPath) > 1 && finalPath[1] == ':' {
			finalPath = strings.ToUpper(finalPath[:1]) + finalPath[1:]
		}

		// Línea de Debug: Muestra la ruta limpia que se está intentando ejecutar
		fmt.Println("Intentando ejecutar con ruta (debug):", finalPath)

		// ******* EJECUCIÓN MEJORADA CON EXPLORER.EXE *******
		// La forma más robusta de hacer que Windows abra un archivo asociado (.bat)
		// sin problemas de sintaxis del shell es usando explorer.exe
		cmd := exec.Command("explorer", finalPath)

		// Intentamos iniciar el proceso
		if err := cmd.Start(); err != nil {
			// Si hay un error, lo mostramos con el path exacto intentado.
			showMessage(w, "Error de Ejecución", fmt.Sprintf("No se pudo iniciar el archivo BAT usando explorer.exe. Compruebe la ruta en la consola. Error: %v", err))
			return
		}

		// Si llega aquí, asumimos que el comando start se lanzó correctamente.
		showMessage(w, "Lanzamiento", "El archivo .bat se ha lanzado correctamente.")
		// *************************
	})
	executeButton.Disable() // Deshabilitado al inicio

	// Botón principal para generar el archivo .bat
	generateButton := widget.NewButton("Generar Archivo .bat", func() {
		success, path := createBatFile(w, selectedExePath) // Usamos selectedExePath
		if success {
			// ** ACTUALIZAR ESTADO DE EJECUCIÓN: Se usará la ruta recién generada **
			pathForExecution = path
			pathEntry.SetText(path) // Actualiza el campo para mostrar la ruta BAT generada
			executeButton.Enable()
			showMessage(w, "Éxito", fmt.Sprintf("¡Archivo .bat creado con éxito!\nAhora puedes ejecutarlo directamente."))
		}
	})

	// Botón para seleccionar el archivo .exe (mantiene la lógica de creación)
	selectExeButton := widget.NewButtonWithIcon("Seleccionar Archivo .exe", theme.FolderOpenIcon(), func() {
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				return
			}
			selectedExePath = read.URI().Path()
			pathEntry.SetText(selectedExePath)

			// Calcular la ruta del .bat esperado
			basePath := strings.TrimSuffix(selectedExePath, filepath.Ext(selectedExePath))
			expectedBatPath := basePath + "_RunAsInvoker.bat"

			// Si el .bat esperado existe, prepáralo para la ejecución
			if _, err := os.Stat(filepath.FromSlash(expectedBatPath)); err == nil {
				pathForExecution = expectedBatPath
				executeButton.Enable()
			} else {
				pathForExecution = ""
				executeButton.Disable()
			}

		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
		fd.Show()
	})

	// Botón para seleccionar un archivo .bat manualmente
	selectBatButton := widget.NewButtonWithIcon("Seleccionar .bat", theme.FileIcon(), func() {
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				return
			}
			batPath := read.URI().Path()
			pathEntry.SetText(batPath)

			// ** ACTUALIZAR ESTADO DE EJECUCIÓN: Se usará la ruta seleccionada **
			pathForExecution = batPath
			executeButton.Enable()
			selectedExePath = "" // Limpiar la ruta del exe para evitar confusiones en la generación

		}, w)

		fd.SetFilter(storage.NewExtensionFileFilter([]string{".bat"}))
		fd.Show()
	})

	// Contenedor de botones de selección
	selectionButtons := container.NewHBox(
		selectExeButton,
		selectBatButton,
	)

	// Contenedor de botones de acción (Generar/Ejecutar)
	actionButtons := container.NewHBox(
		layout.NewSpacer(), // Empujar los botones hacia el centro
		generateButton,
		executeButton,
		layout.NewSpacer(), // Empujar los botones hacia el centro
	)

	// Contenido principal
	content := container.New(layout.NewVBoxLayout(),
		widget.NewLabel("1. Selecciona el archivo .exe para generar (o .bat para ejecutar):"),
		container.New(layout.NewBorderLayout(nil, nil, nil, selectionButtons),
			selectionButtons, pathEntry),
		widget.NewSeparator(),
		actionButtons,
	)

	w.SetContent(container.NewPadded(content))
	w.ShowAndRun()
}
