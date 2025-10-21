// Aplicación de línea de comandos (CLI) para generar archivos .bat
package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Constante que define el contenido de la plantilla del archivo .bat
const batContent = `Set __COMPAT_LAYER=RunAsInvoker
Start "" "%s"`

// logAndExit imprime un mensaje de error y termina el programa.
func logAndExit(message string) {
	fmt.Fprintln(os.Stderr, "Error:", message)
	os.Exit(1)
}

// createBatFile contiene la lógica de generación.
func createBatFile(exePath string) (string, error) {
	// 1. Verificar si la ruta es válida y tiene la extensión .exe
	if exePath == "" {
		return "", fmt.Errorf("la ruta del ejecutable no puede estar vacía")
	}
	if filepath.Ext(exePath) != ".exe" {
		return "", fmt.Errorf("el archivo debe ser un ejecutable (.exe)")
	}

	dirPath := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	// 2. Definir la ruta del nuevo archivo .bat
	baseName := exeName[:len(exeName)-len(filepath.Ext(exeName))]
	batFileName := fmt.Sprintf("%s_launcher.bat", baseName)
	batPath := filepath.Join(dirPath, batFileName)

	// 3. Formatear el contenido
	content := fmt.Sprintf(batContent, exeName)

	// 4. Escribir el contenido en el nuevo archivo .bat
	err := ioutil.WriteFile(batPath, []byte(content), 0755)
	if err != nil {
		return "", fmt.Errorf("no se pudo crear el archivo .bat: %v", err)
	}

	return batPath, nil
}

func main() {
	// La aplicación CLI espera exactamente un argumento (la ruta del .exe)
	if len(os.Args) != 2 {
		fmt.Println("Uso: bat_cli_app.exe <ruta_al_ejecutable.exe>")
		fmt.Println("Ejemplo: bat_cli_app.exe C:\\Juegos\\MiJuego.exe")
		os.Exit(0)
	}

	exePath := os.Args[1]

	batPath, err := createBatFile(exePath)
	if err != nil {
		logAndExit(err.Error())
	}

	fmt.Println("-------------------------------------------------------")
	fmt.Println("¡Archivo .bat creado con éxito!")
	fmt.Printf("Ruta de salida: %s\n", batPath)
	fmt.Println("-------------------------------------------------------")
}
