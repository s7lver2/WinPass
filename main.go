// Paquete principal de la aplicación
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const VERSION = "WinPass v3.1"

// Características activables por el patcher
const DEBUG = false                 // debug
const SECURE_DELETE = false // Borrado seguro del .bat después de ejecución
const SECURE_DELETE_SECONDS_DELAY = 10
const SECURE_DELETE_SECONDS_RETRY_DELAY = 3
const SECURE_DELETE_RETRY_ATTEMPTS = 5

// Plantilla del .bat
const batContent = `@echo off
Set __COMPAT_LAYER=RunAsInvoker
Start "" "%%~dp0%s"`

// Variable de estado para la ruta final a ejecutar
var pathForExecution string = ""

// showMessage muestra un cuadro de diálogo simple
func showMessage(w fyne.Window, title string, message string) {
	d := dialog.NewInformation(title, message, w)
	d.Show()
}

// secureDelete borra el archivo y sobrescribe el espacio en disco (solo si FEATURE_SECURE_DELETE = true)
// Incluye reintentos para manejar "archivo en uso" temporalmente
func secureDelete(filePath string, w fyne.Window) {
    const retryDelay = SECURE_DELETE_SECONDS_RETRY_DELAY * time.Second // Espera 3 segundos entre intentos

    var err error

    // Intentar borrar el archivo con reintentos
    for attempt := 1; attempt <= SECURE_DELETE_RETRY_ATTEMPTS; attempt++ {
        err = os.Remove(filePath)
        if err == nil {
            // Borrado exitoso
            break
        }

        // Si el error es "en uso", esperar y reintentar
        if strings.Contains(strings.ToLower(err.Error()), "used by another process") ||
           strings.Contains(strings.ToLower(err.Error()), "being used") {
            if DEBUG {
                showMessage(w, "DEBUG", fmt.Sprintf("Archivo en uso (intento %d/%d). Esperando %d segundos...", attempt, SECURE_DELETE_RETRY_ATTEMPTS, int(retryDelay.Seconds())))
            }
            time.Sleep(retryDelay)
            continue
        }

        // Otro error (no es "en uso"), salir del bucle
        showMessage(w, "Advertencia", fmt.Sprintf("No se pudo borrar el .bat: %v", err))
        return
    }

    // Si después de reintentos sigue fallando
    if err != nil {
        showMessage(w, "Advertencia", fmt.Sprintf("No se pudo borrar el .bat después de %d intentos.\nInténtalo manualmente.", SECURE_DELETE_RETRY_ATTEMPTS))
        return
    }

    // Borrado exitoso: ahora sobrescribir si la opción está activada
    if !SECURE_DELETE {
        if DEBUG {
            showMessage(w, "DEBUG", "Archivo .bat eliminado (sin sobrescritura)")
        }
        return
    }

    // Sobrescritura segura con cipher /w
    dir := filepath.Dir(filePath)
    cmd := exec.Command("cipher", "/w:"+dir)
    cmd.Stdout = os.Stdout // Opcional: ver progreso en consola
    cmd.Stderr = os.Stderr

    err = cmd.Run()
    if err != nil {
        showMessage(w, "Advertencia", fmt.Sprintf("cipher /w falló: %v (archivo ya borrado)", err))
    } else {
        if DEBUG {
            showMessage(w, "DEBUG", "Archivo .bat eliminado y espacio sobrescrito con cipher /w")
        }
    }
}

// createBatFile genera el archivo .bat
func createBatFile(w fyne.Window, exePath string) (bool, string) {
	if exePath == "" || filepath.Ext(exePath) != ".exe" {
		showMessage(w, "ERROR", "Selecciona un archivo .exe válido")
		return false, ""
	}

	dirPath := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)
	baseName := exeName[:len(exeName)-len(filepath.Ext(exePath))]
	batFileName := fmt.Sprintf("%s_Payload.bat", baseName)
	batPath := filepath.Join(dirPath, batFileName)

	content := fmt.Sprintf(batContent, exeName)

	err := ioutil.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		showMessage(w, "ERROR", fmt.Sprintf("No se pudo crear el payload: %v", err))
		return false, ""
	}

	return true, batPath
}

// createGeneratorTab (igual que antes, solo con mensajes en español)
func createGeneratorTab(w fyne.Window, pathEntry *widget.Entry, executeButton *widget.Button) fyne.CanvasObject {
	var selectedExePath string
	savePathLabel := widget.NewLabel("Ruta de guardado: No seleccionada")

	generateButton := widget.NewButton("Generar Payload", func() {
		success, path := createBatFile(w, selectedExePath)
		if success {
			pathForExecution = path
			pathEntry.SetText(path)
			executeButton.Enable()
			showMessage(w, "Éxito", "Payload generado correctamente. Ve a Ejecutar para lanzarlo.")
		}
	})

	selectExeButton := widget.NewButtonWithIcon("Seleccionar .exe", theme.FileIcon(), func() {
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				return
			}
			selectedExePath = read.URI().Path()
			basePath := strings.TrimSuffix(selectedExePath, filepath.Ext(selectedExePath))
			expectedBatPath := basePath + "_Payload.bat"
			savePathLabel.SetText(fmt.Sprintf("Se guardará en:\n%s", filepath.FromSlash(expectedBatPath)))

			if _, err := os.Stat(filepath.FromSlash(expectedBatPath)); err == nil {
				pathForExecution = expectedBatPath
				pathEntry.SetText(expectedBatPath)
				executeButton.Enable()
			} else {
				pathForExecution = ""
				pathEntry.SetText(selectedExePath)
				executeButton.Disable()
			}
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
		fd.Show()
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("WinPass | Generar Payload", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Paso 1: Selecciona el .exe"),
		selectExeButton,
		widget.NewSeparator(),
		widget.NewLabel("Paso 2: Genera el payload"),
		generateButton,
		widget.NewSeparator(),
		savePathLabel,
	)
}

// createExecutionTab (igual)
func createExecutionTab(w fyne.Window, pathEntry *widget.Entry, executeButton *widget.Button) fyne.CanvasObject {
	selectBatButton := widget.NewButtonWithIcon("Seleccionar Payload (.bat)", theme.DocumentIcon(), func() {
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				return
			}
			batPath := read.URI().Path()
			pathEntry.SetText(batPath)
			pathForExecution = batPath
			executeButton.Enable()
		}, w)
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".bat"}))
		fd.Show()
	})

	return container.NewVBox(
		widget.NewLabelWithStyle("WinPass | Ejecutar Payload", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Paso 1: Selecciona el _Payload.bat"),
		selectBatButton,
		widget.NewSeparator(),
		widget.NewLabel("Paso 2: Ejecuta el payload"),
		executeButton,
		widget.NewSeparator(),
		widget.NewLabel("Payload cargado:"),
		pathEntry,
	)
}

func main() {
	if DEBUG {
		fmt.Println("DEBUG activado")
		fmt.Printf("SECURE_DELETE: %v\n", SECURE_DELETE)
	}

	a := app.New()
	w := a.NewWindow(VERSION)
	w.Resize(fyne.NewSize(750, 400))
	w.SetFixedSize(true)

	pathEntry := widget.NewEntry()
	pathEntry.PlaceHolder = "Ruta del payload a ejecutar"
	pathEntry.Disable()

	// Sincronización para esperar ejecución (opcional, pero útil)
	var wg sync.WaitGroup

	executeButton := widget.NewButtonWithIcon("Ejecutar Payload", theme.MediaPlayIcon(), func() {
		if pathForExecution == "" {
			showMessage(w, "ERROR", "Selecciona un payload válido")
			return
		}

		batPath := pathForExecution

		// Ejecutar el .bat sin ventana visible
		cmd := exec.Command("cmd", "/c", "start", "/b", batPath)
		cmd.Dir = filepath.Dir(batPath)

		err := cmd.Start()
		if err != nil {
			showMessage(w, "ERROR", fmt.Sprintf("No se pudo iniciar el payload: %v", err))
			return
		}

		// Esperar en goroutine para no bloquear GUI
		wg.Add(1)
		// Esperar en goroutine para no bloquear la GUI
    	go func() {
        	err := cmd.Wait()
        	if err != nil {
        	    showMessage(w, "Advertencia", fmt.Sprintf("El payload terminó con error: %v", err))
        	}

        	// Esperar 10 segundos antes de borrar (ajusta el tiempo si quieres)
        	showMessage(w, "Información", fmt.Sprintf("Payload ejecutado. Esperando %d segundos antes de eliminar el payload", SECURE_DELETE_SECONDS_DELAY))
        	time.Sleep(SECURE_DELETE_SECONDS_DELAY * time.Second)

        	// Borrar de forma segura (o normal según la constante)
        	secureDelete(batPath, w)
    	}()

		showMessage(w, "Éxito", "Payload ejecutado. Se eliminará automáticamente al terminar.")
	})
	executeButton.Disable()

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("", theme.ContentAddIcon(),
			container.NewPadded(createGeneratorTab(w, pathEntry, executeButton)),
		),
		container.NewTabItemWithIcon("", theme.MediaPlayIcon(),
			container.NewPadded(createExecutionTab(w, pathEntry, executeButton)),
		),
	)
	tabs.SetTabLocation(container.TabLocationLeading)

	w.SetContent(tabs)
	w.ShowAndRun()
}