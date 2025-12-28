// Paquete principal del Patcher
package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

// Define las versiones disponibles (agrega más según necesites)
var versions = []string{"v3.1", "v3.0"}

// Define las opciones por versión (agrega o modifica según las features que quieras patchear)
func getOptionsForVersion(version string) []string {
	switch version {
	case "v3.1":
		return []string{"Enable Debug", "Enable Feature A"}
	case "v3.0":
		return []string{"Enable Feature B"}
	default:
		return []string{}
	}
}

// Mapeo de opciones a nombres de variables en el código (asumiendo const VAR = false en el main.go original)
var optionToVar = map[string]string{
	"Enable Debug":    "DEBUG",
	"Enable Feature A": "FEATURE_A",
	"Enable Feature B": "FEATURE_B",
}

// Mapeo de versión a URL del ZIP (reemplaza con URLs reales, e.g., desde GitHub: https://github.com/tu-repo/winpass/archive/refs/tags/v3.1.zip)
var versionToURL = map[string]string{
	"v3.1": "http://localhost/winpass-v3.1.zip", // Reemplaza con URL real
	"v3.0": "http://localhost/winpass-v3.0.zip", // Reemplaza con URL real
}

func performPatch(version string, selectedOpts []string, w fyne.Window) error {
    url, ok := versionToURL[version]
    if !ok {
        return fmt.Errorf("no hay URL definida para la versión %s", version)
    }

    // Paso 1: Descarga
    dialog.ShowInformation("Descargando", "Descargando ZIP de "+url, w)
    resp, err := http.Get(url)
    if err != nil {
        return fmt.Errorf("fallo al descargar: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("respuesta HTTP inválida: %s (código %d)", resp.Status, resp.StatusCode)
    }

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("fallo al leer ZIP: %v", err)
    }

    // Paso 2: Crea directorio temporal
    tempDir, err := os.MkdirTemp("", "winpass-patcher-*")
    if err != nil {
        return fmt.Errorf("fallo al crear carpeta temporal: %v", err)
    }
    defer os.RemoveAll(tempDir) // Limpia al final

    // Paso 3: Descomprime
    reader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
    if err != nil {
        return fmt.Errorf("fallo al leer ZIP: %v", err)
    }

    var sourceDir string
    for _, file := range reader.File {
        path := filepath.Join(tempDir, file.Name)
        if file.FileInfo().IsDir() {
            os.MkdirAll(path, os.ModePerm)
            if sourceDir == "" && strings.HasSuffix(file.Name, "/") && strings.Count(file.Name, "/") == 1 {
                sourceDir = path
            }
            continue
        }

        f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
        if err != nil {
            return fmt.Errorf("fallo al crear archivo %s: %v", file.Name, err)
        }
        rc, err := file.Open()
        if err != nil {
            f.Close()
            return fmt.Errorf("fallo al abrir archivo ZIP %s: %v", file.Name, err)
        }
        _, err = io.Copy(f, rc)
        rc.Close()
        f.Close()
        if err != nil {
            return fmt.Errorf("fallo al escribir %s: %v", file.Name, err)
        }
    }

    if sourceDir == "" {
        sourceDir = tempDir
    }

    mainPath := filepath.Join(sourceDir, "main.go")
    if _, err := os.Stat(mainPath); os.IsNotExist(err) {
        return fmt.Errorf("no se encontró main.go en %s", sourceDir)
    }

    // Paso 4: Modificar main.go
    data, err := os.ReadFile(mainPath)
    if err != nil {
        return fmt.Errorf("fallo al leer main.go: %v", err)
    }

    code := string(data)
    modified := false
    for _, opt := range selectedOpts {
        varName, exists := optionToVar[opt]
        if !exists {
            continue
        }
        old := fmt.Sprintf("const %s = false", varName)
        new := fmt.Sprintf("const %s = true", varName)
        if strings.Contains(code, old) {
            code = strings.Replace(code, old, new, 1)
            modified = true
        }
    }

    if !modified && len(selectedOpts) > 0 {
        return fmt.Errorf("ninguna opción seleccionada se pudo aplicar (¿las constantes existen en el código?)")
    }

    if err := os.WriteFile(mainPath, []byte(code), 0644); err != nil {
        return fmt.Errorf("fallo al escribir main.go modificado: %v", err)
    }
    // Paso 4: resolver dependencias
	dialog.ShowInformation("Preparando", "Resolviendo dependencias Go...", w)

	// 1. go mod tidy (genera/completa go.sum)
	cmd := exec.Command("go", "mod", "tidy", "-v")
	cmd.Dir = sourceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
    	return fmt.Errorf("fallo al ejecutar go mod tidy: %v\nSalida:\n%s", err, output)
	}

    // Paso 5: Compilar
	dialog.ShowInformation("Compilando", "Compilando la versión patcheada...", w)
    cmd = exec.Command("go", "build", "-o", "patched.exe")
    cmd.Dir = sourceDir
    output, err = cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("fallo al compilar: %v\nSalida:\n%s", err, output)
    }

	// Paso 6: Mover el fichero
    srcPath := filepath.Join(sourceDir, "patched.exe")
    destPath := filepath.Join(".", fmt.Sprintf("winpass-%s-patched.exe", version))

    srcFile, err := os.Open(srcPath)
    if err != nil {
        return fmt.Errorf("fallo al abrir el .exe generado: %v", err)
    }
    defer srcFile.Close()

    destFile, err := os.Create(destPath)
    if err != nil {
        return fmt.Errorf("fallo al crear el .exe destino: %v", err)
    }
    defer destFile.Close()

    if _, err := io.Copy(destFile, srcFile); err != nil {
        return fmt.Errorf("fallo al copiar el .exe: %v", err)
    }

    // Dar permisos de ejecución (útil en Linux)
    if err := os.Chmod(destPath, 0755); err != nil {
        return fmt.Errorf("fallo al dar permisos de ejecución: %v", err)
    }

    // Mensaje de éxito (opcional, agrégalo antes del return nil)
    dialog.ShowInformation("Éxito", "El archivo patcheado ha sido creado: "+destPath, w)

    return nil
}

func main() {
	a := app.New()
	w := a.NewWindow("WinPass Patcher")
	w.Resize(fyne.NewSize(500, 400))

	// Selección de versión
	versionSelect := widget.NewSelect(versions, nil)

	// Contenedor dinámico para opciones
	optionsContainer := container.NewVBox(widget.NewLabel("Selecciona una versión para ver opciones"))

	// Mapa de checkboxes
	checkboxes := make(map[string]*widget.Check)

	// Actualizar opciones al cambiar versión
	versionSelect.OnChanged = func(selectedVersion string) {
		opts := getOptionsForVersion(selectedVersion)
		optionsContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Opciones disponibles:")}
		checkboxes = make(map[string]*widget.Check)
		for _, opt := range opts {
			check := widget.NewCheck(opt, nil)
			checkboxes[opt] = check
			optionsContainer.Add(check)
		}
		optionsContainer.Refresh()
	}

	// Botón de patch
	patchButton := widget.NewButtonWithIcon("Patch", theme.ConfirmIcon(), func() {
		version := versionSelect.Selected
		if version == "" {
			dialog.ShowError(fmt.Errorf("Selecciona una versión"), w)
			return
		}

		selectedOpts := []string{}
		for opt, check := range checkboxes {
			if check.Checked {
				selectedOpts = append(selectedOpts, opt)
			}
		}

		err := performPatch(version, selectedOpts, w)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		dialog.ShowInformation("Éxito", "El .exe patcheado ha sido creado en el directorio actual.", w)
	})

	// Contenido principal
	content := container.NewVBox(
		widget.NewLabelWithStyle("WinPass Patcher", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Selecciona la versión:"),
		versionSelect,
		widget.NewSeparator(),
		optionsContainer,
		widget.NewSeparator(),
		patchButton,
	)

	w.SetContent(container.NewPadded(content))
	w.ShowAndRun()
}