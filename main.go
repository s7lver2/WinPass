// Paquete principal de la aplicación
package main

import (
	"fmt"       // Para formatear cadenas (como el contenido del archivo .bat)
	"io/ioutil" // Para escribir archivos (el archivo .bat)
	// Para interactuar con el sistema de archivos
	"path/filepath" // Para manejar rutas de archivos

	// Importaciones de Fyne para la GUI
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage" // Importación necesaria para el filtro de archivos
	"fyne.io/fyne/v2/theme"   // Importación necesaria para usar los íconos del tema
	"fyne.io/fyne/v2/widget"
)

// Constante que define el contenido de la plantilla del archivo .bat
// El placeholder %s será reemplazado por el nombre del archivo .exe
const batContent = `Set __COMPAT_LAYER=RunAsInvoker
Start "" "%s"`

// Definición del icono de la aplicación.
// La herramienta 'fyne package' busca esta variable globalmente para el icono del .exe.
var Icon = theme.FolderOpenIcon()

// showMessage muestra un cuadro de diálogo simple para notificaciones
func showMessage(w fyne.Window, title string, message string) {
	d := dialog.NewInformation(title, message, w)
	d.Show()
}

// createBatFile es la lógica principal para generar el archivo .bat
func createBatFile(w fyne.Window, exePath string) {
	// 1. Verificar si la ruta del .exe es válida y tiene la extensión .exe
	if exePath == "" {
		showMessage(w, "Error", "Por favor, selecciona un archivo .exe válido.")
		return
	}
	if filepath.Ext(exePath) != ".exe" {
		showMessage(w, "Error", "El archivo seleccionado debe ser un ejecutable (.exe).")
		return
	}

	// 2. Obtener la ruta del directorio y el nombre del archivo sin ruta
	dirPath := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	// 3. Definir la ruta del nuevo archivo .bat
	batFileName := fmt.Sprintf("%s_launcher.bat", exeName[:len(exeName)-len(filepath.Ext(exeName))])
	batPath := filepath.Join(dirPath, batFileName)

	// 4. Formatear el contenido del archivo .bat
	// Usamos el nombre del archivo .exe (exeName) para el contenido del bat
	content := fmt.Sprintf(batContent, exeName)

	// 5. Escribir el contenido en el nuevo archivo .bat
	err := ioutil.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		showMessage(w, "Error de Archivo", fmt.Sprintf("No se pudo crear el archivo .bat:\n%v", err))
		return
	}

	// 6. Éxito
	showMessage(w, "Éxito", fmt.Sprintf("¡Archivo .bat creado con éxito!\n\nNombre: %s\nUbicación: %s", batFileName, dirPath))
}

func main() {
	// Inicializa la aplicación Fyne
	a := app.New()

	// ******* CAMBIO: a.SetIcon(Icon) ya no es necesario; la variable global Icon lo hace automáticamente. *******
	// Esta línea ha sido eliminada para evitar conflictos con 'fyne package'.

	w := a.NewWindow("Generador BAT (RunAsInvoker)")
	// Ventana ajustada a un tamaño más grande: 750x400
	w.Resize(fyne.NewSize(750, 400))
	w.SetFixedSize(true)

	// Variable para almacenar la ruta del archivo seleccionado
	var selectedPath string

	// Campo de texto para mostrar la ruta del archivo .exe seleccionado
	pathEntry := widget.NewEntry()
	pathEntry.PlaceHolder = "Ruta del archivo .exe..."
	pathEntry.Disable() // No se permite la edición manual

	// Botón principal para seleccionar el archivo .exe
	// CORRECCIÓN: Usamos theme.FolderOpenIcon() ya que es el ícono estándar de selección
	selectButton := widget.NewButtonWithIcon("Seleccionar Archivo .exe", theme.FolderOpenIcon(), func() {
		// Crea un diálogo de selección de archivo
		fd := dialog.NewFileOpen(func(read fyne.URIReadCloser, err error) {
			if err != nil || read == nil {
				// El usuario canceló o hubo un error
				return
			}

			// Si se selecciona un archivo, actualiza la ruta
			selectedPath = read.URI().Path()
			pathEntry.SetText(selectedPath)
		}, w)

		// Opcional: Establece un filtro para mostrar solo archivos .exe
		fd.SetFilter(storage.NewExtensionFileFilter([]string{".exe"}))
		fd.Show()
	})

	// Botón para generar el archivo .bat
	generateButton := widget.NewButton("Generar Archivo .bat", func() {
		createBatFile(w, selectedPath)
	})

	// Contenedor principal con un diseño de cuadrícula para centrar los elementos
	content := container.New(layout.NewVBoxLayout(),
		widget.NewLabel("1. Selecciona el archivo .exe:"),
		container.New(layout.NewBorderLayout(nil, nil, nil, selectButton),
			selectButton, pathEntry),
		widget.NewSeparator(),
		container.New(layout.NewCenterLayout(), generateButton),
	)

	// Establecer el contenido de la ventana
	w.SetContent(container.NewPadded(content))

	w.ShowAndRun()
}
