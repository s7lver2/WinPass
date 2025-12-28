// Paquete principal del Patcher
package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
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

// Estructuras para el JSON de configuración
type PatchOption struct {
	Label   string `json:"label"`
	VarName string `json:"var_name"`
}

type VersionConfig struct {
	Name    string        `json:"name"`
	URL     string        `json:"url"`
	Options []PatchOption `json:"options"`
}

type Config struct {
	Versions     []VersionConfig `json:"versions"`
	OutputPrefix string          `json:"output_prefix"`
	OutputSuffix string          `json:"output_suffix"`
}

// Variable global para la configuración cargada
var config Config

// loadConfig carga el archivo config.json
func loadConfig() error {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return fmt.Errorf("no se pudo leer config.json: %v", err)
	}
	return json.Unmarshal(data, &config)
}

// getVersionByName busca la configuración de una versión por nombre
func getVersionByName(name string) (VersionConfig, bool) {
	for _, v := range config.Versions {
		if v.Name == name {
			return v, true
		}
	}
	return VersionConfig{}, false
}

// copyFile copia un archivo (evita problemas de cross-device link)
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	dest, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dest.Close()

	_, err = io.Copy(dest, source)
	if err != nil {
		return err
	}

	// Dar permisos de ejecución (útil en Linux)
	return os.Chmod(dst, 0755)
}

func performPatch(versionName string, selectedLabels []string, w fyne.Window) error {
	ver, ok := getVersionByName(versionName)
	if !ok {
		return fmt.Errorf("versión no encontrada: %s", versionName)
	}

	// Paso 1: Descarga
	dialog.ShowInformation("Descargando", "Descargando ZIP de "+ver.URL, w)
	resp, err := http.Get(ver.URL)
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

	// Ajusta la ruta de main.go según la estructura del ZIP (ejemplo común)
	// Si el ZIP tiene una carpeta raíz como "WinPass-v3.1/", cámbialo aquí
	mainPath := filepath.Join(sourceDir, "main.go")
	if _, err := os.Stat(mainPath); os.IsNotExist(err) {
		// Intento alternativo: buscar en subcarpeta
		mainPath = filepath.Join(sourceDir, "WinPass-"+versionName, "main.go")
		if _, err := os.Stat(mainPath); os.IsNotExist(err) {
			return fmt.Errorf("no se encontró main.go en %s", sourceDir)
		}
	}

	// Paso 4: Modificar main.go
	data, err := os.ReadFile(mainPath)
	if err != nil {
		return fmt.Errorf("fallo al leer main.go: %v", err)
	}

	code := string(data)
	modified := false
	for _, label := range selectedLabels {
		for _, opt := range ver.Options {
			if opt.Label == label {
				old := fmt.Sprintf("const %s = false", opt.VarName)
				new := fmt.Sprintf("const %s = true", opt.VarName)
				if strings.Contains(code, old) {
					code = strings.Replace(code, old, new, 1)
					modified = true
				}
			}
		}
	}

	if !modified && len(selectedLabels) > 0 {
		return fmt.Errorf("ninguna opción seleccionada se pudo aplicar (¿las constantes existen en el código?)")
	}

	if err := os.WriteFile(mainPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("fallo al escribir main.go modificado: %v", err)
	}

	// Paso 5: Resolver dependencias (go mod tidy)
	dialog.ShowInformation("Preparando", "Resolviendo dependencias Go...", w)
	cmd := exec.Command("go", "mod", "tidy", "-v")
	cmd.Dir = sourceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fallo al ejecutar go mod tidy: %v\nSalida:\n%s", err, string(output))
	}

	// Paso 6: Compilar
	dialog.ShowInformation("Compilando", "Compilando la versión patcheada...", w)
	cmd = exec.Command("go", "build", "-ldflags=-s -w -H=windowsgui", "-o", "patched.exe")
	cmd.Dir = sourceDir
	output, err = cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("fallo al compilar: %v\nSalida:\n%s", err, string(output))
	}

	// Paso 7: Copiar el .exe generado al directorio actual
	srcPath := filepath.Join(sourceDir, "patched.exe")
	destPath := fmt.Sprintf("%s%s%s", config.OutputPrefix, versionName, config.OutputSuffix)

	if err := copyFile(srcPath, destPath); err != nil {
		return fmt.Errorf("fallo al copiar el .exe: %v", err)
	}

	dialog.ShowInformation("Éxito", "El archivo patcheado ha sido creado: "+destPath, w)
	return nil
}

func main() {
	if err := loadConfig(); err != nil {
		dialog.ShowError(err, nil) // Mostrar error en GUI si falla
		os.Exit(1)
	}

	a := app.New()
	w := a.NewWindow("WinPass Patcher")
	w.Resize(fyne.NewSize(500, 400))

	// Lista de nombres de versiones
	versionNames := make([]string, len(config.Versions))
	for i, v := range config.Versions {
		versionNames[i] = v.Name
	}

	versionSelect := widget.NewSelect(versionNames, nil)

	// Contenedor dinámico de opciones
	optionsContainer := container.NewVBox(widget.NewLabel("Selecciona una versión"))
	checkboxes := make(map[string]*widget.Check)

	versionSelect.OnChanged = func(selected string) {
		ver, _ := getVersionByName(selected)
		optionsContainer.Objects = []fyne.CanvasObject{widget.NewLabel("Opciones disponibles:")}
		checkboxes = make(map[string]*widget.Check)
		for _, opt := range ver.Options {
			ch := widget.NewCheck(opt.Label, nil)
			checkboxes[opt.Label] = ch
			optionsContainer.Add(ch)
		}
		optionsContainer.Refresh()
	}

	// Botón de patch
	patchButton := widget.NewButtonWithIcon("Patch", theme.ConfirmIcon(), func() {
		version := versionSelect.Selected
		if version == "" {
			dialog.ShowError(fmt.Errorf("selecciona una versión"), w)
			return
		}

		var selected []string
		for label, ch := range checkboxes {
			if ch.Checked {
				selected = append(selected, label)
			}
		}

		err := performPatch(version, selected, w)
		if err != nil {
			dialog.ShowError(err, w)
		} else {
			dialog.ShowInformation("Éxito", "Parcheado creado correctamente", w)
		}
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