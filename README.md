# WinPass | Windows Admin Bypass

![bat_generator_banner](.github/ISSUE_TEMPLATE/animesher.com_pixel-pixel-gif-gif-2066449.gif)

<!--
Multi-language README support
-->
<!-- ALL-CONTRIBUTORS-BADGE:START - Do not remove or modify this section -->
[![All Contributors](https://img.shields.io/badge/all_contributors-1-orange.svg?style=flat-square)](#contributors-)
<!-- ALL-CONTRIBUTORS-BADGE:END -->
[![es](https://img.shields.io/badge/lang-es-yellow.svg)](Source/docs/README.es.md)
[![en](https://img.shields.io/badge/lang-en-blue.svg)](Source/docs/README.en.md)

<div align="center">

<br>

<a href="#requirements"><kbd> <br>📋 Requirements <br> </kbd></a>&ensp;&ensp;
<a href="#versions"><kbd> <br>🔄 Versions <br> </kbd></a>&ensp;&ensp;
<a href="#compilation"><kbd> <br>⚙️ Compilation <br> </kbd></a>&ensp;&ensp;
<a href="#usage"><kbd> <br>🚀 Usage <br> </kbd></a>&ensp;&ensp;
<a href="#troubleshooting"><kbd> <br>🔧 Troubleshooting <br> </kbd></a>&ensp;&ensp;
<a href="CONTRIBUTING.md"><kbd> <br>🤝 Contributing <br> </kbd></a>&ensp;&ensp;
<a href="https://github.com/s7lver2/WinPass/issues"><kbd> <br>🐛 Issues <br> </kbd></a>&ensp;&ensp;
<a href="https://github.com/s7lver2/WinPass/releases"><kbd> <br>💿 Releases<br> </kbd></a>

</div><br><br>

<div align="center">

<table>
  <tr>
    <td align="center">
      <img src="https://github.com/s7lver2/WinPass/blob/3b530240b16d67d05102609c0432d65640c73ffb/.github/ISSUE_TEMPLATE/w7.png" alt="Windows 7 Logo" width="50" height="50" /><br>
      <sub><strong>Windows 7</strong></sub>
    </td>
    <td align="center">
      <img src="https://github.com/s7lver2/WinPass/blob/3b530240b16d67d05102609c0432d65640c73ffb/.github/ISSUE_TEMPLATE/w10.png" alt="Windows 10 Logo" width="50" height="50" /><br>
      <sub><strong>Windows 10</strong></sub>
    </td>
    <td align="center">
      <img src="https://github.com/s7lver2/WinPass/blob/3b530240b16d67d05102609c0432d65640c73ffb/.github/ISSUE_TEMPLATE/w11.png" alt="Windows 11 Logo" width="50" height="50" /><br>
      <sub><strong>Windows 11</strong></sub>
    </td>
  </tr>
</table>

</div>

Check this out for the full note:
[Check my latest projects!](https://github.com/s7lver2?tab=repositories)

<br>

<a id="requirements"></a>
<img src="https://readme-typing-svg.herokuapp.com?font=Lexend+Giga&size=25&pause=1000&color=CCA9DD&vCenter=true&width=435&height=25&lines=REQUIREMENTS" width="450"/>

---
This is a GUI app builded with Go for Windows than generates ´.bat´ files with the compatibility layer ´RunAsInvoker´ for run executables ´.exe´ without elevating privileadges UAC. It is compatible with almost every windows system, like **Windows 7**, **Windows 10** and **Windows 11**.

> [!IMPORTANT]
> I am not responsible for the uses of this tool, and if you do something ilegal or bad with it, it is your responasbility
---

<a id="versions"></a>
<img src="https://readme-typing-svg.herokuapp.com?font=Lexend+Giga&size=25&pause=1000&color=CCA9DD&vCenter=true&width=435&height=25&lines=VERSIONS" width="450"/>

---
This repository includes 2 versions of the project, each one of them optimized for diferent enviroments and compabilities. We strongly recommend the usage of the **Main** version

1. **Main (parent directory: `../`)** GUI version made with Fyne (modern UI with tabs). Stable support for **Windows 10/11**. experimental support for **Windows 7/8**, but with graphical aceleration requirement (OpenGL/DirectX). Ideal for new users than dont really know too much about it

2. **Legacy (actual directory: `Legacy/`)**: CLI version with native Windows dialogs and manual fallback. It supports every windows from **Windows 7 SP1+** without dependences requirements, like MinGW. Usefull for legacy or old systems without GUI


Directories:
```shell
cd ..      # For Main (main, Win10/11)
cd Legacy  # For Legacy (secondary, Win7)
```

> [!NOTE]
> The main version uses its own `go.mod` file. Compile separately for dont have dependences conflicts
---

<a id="compilation"></a>
<img src="https://readme-typing-svg.herokuapp.com?font=Lexend+Giga&size=25&pause=1000&color=CCA9DD&vCenter=true&width=435&height=25&lines=COMPILATION" width="450"/>

---

> [!IMPORTANT]
> Compilate with Go 1.20.14 for windows 7 support (newer versions fail because `bcryptprimitives.dll`).

> [!CAUTION]
> for compile la legacy version, necesitas Go 1.20.14 y `golang.org/x/sys@v0.7.0`.

> [!TIP]
> Usa `goenv` para manejar múltiples versiones de Go sin conflictos.

### Compilación de Main (Win10/11, Experimental Win8/7) - Principal

Se recomienda encarecidamente compilar desde alguna distribución de linux, en el caso de windows, usa WSL [mas info aqui](https://google.com)

#### Preinstalación
1. Instala MinGW:
   ```shell
   sudo apt update && sudo apt install -y gcc-mingw-w64-x86-64 g++-mingw-w64-x86-64 gcc-mingw-w64-i686 g++-mingw-w64-i686 python3 git curl build-essential
   ```

2. Instala goenv:
   ```shell
   git clone https://github.com/syndbg/goenv.git ~/.goenv
   echo 'export GOENV_ROOT="$HOME/.goenv"' >> ~/.bashrc
   echo 'command -v goenv >/dev/null || export PATH="$GOENV_ROOT/bin:$PATH"' >> ~/.bashrc
   echo 'eval "$(goenv init -)"' >> ~/.bashrc
   source ~/.bashrc
   goenv install $(goenv install -l | grep -v rc | tail -1)  # Versión más reciente para Main
   goenv install 1.20.14  # Para Legacy
   ```
#### Compilación Main

3. Navega al directorio padre:
   ```shell
   cd ~/WinPass  # Directorio de Main
   ```

4. Prepara Dependencias:
   ```shell
   go mod tidy  # Resuelve dependencias de Fyne
   ```

5. Compila (requiere CGO para Fyne/OpenGL):
   ```shell
   goenv shell $(goenv install -l | grep -v rc | tail -1)  # Usa la versión más reciente
   GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_ENABLED=1 go build -ldflags="-s -w" -o BAT_GUI.exe main.go
   ```
   - Para 32-bit: Cambia a `i686-w64-mingw32-gcc` y `GOARCH=386`.
   - **Experimental Win7/8**: Prueba con `-tags no_opengl` si hay issues gráficos, pero soporte no garantizado.

### Compilación de Legacy

3. Navega y prepara:
   ```shell
   goenv shell 1.20.14 # activa goenv 1.20.14
   cd ~/WinPass/Legacy
   go mod init WinPass
   go get golang.org/x/sys@v0.7.0
   ```

4. Compila:
   ```shel
   GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o TEST.exe main.go
   ```

**Binario Resultante:** `BAT_GUI.exe` (Main, ~10-15 MB con Fyne) o `TEST.exe` (Legacy, ~2-3 MB).

> [!NOTE]
> Para recuperar los binarios ya compilados, en el caso de que se use **WSL**, se recomienda usar `python -m http.server 80` para poder moverlos facilmente.

---

<a id="usage"></a>
<img src="https://readme-typing-svg.herokuapp.com?font=Lexend+Giga&size=25&pause=1000&color=CCA9DD&vCenter=true&width=435&height=25&lines=USAGE" width="450"/>

---

### Uso de Main (GUI) - Principal
Ejecuta `./Winpass.exe` para abrir la interfaz gráfica con pestañas (Generador y Ejecutar). Selecciona archivos vía diálogos o manual. Recomendado para uso diario en Win10/11.

### Uso de Legacy (CLI) - Secundaria
1. Ejecuta:
   ```cmd
   ./TEST.exe
   ```

2. Menú:
   - **1. Generar BAT**: Ingresa ruta de `.exe` (ej: `C:\Juego\MiApp.exe`). Crea `MiApp_RunAsInvoker.bat`.
   - **2. Ejecutar BAT**: Ingresa ruta de `.bat`.
   - **3. Salir**.

**Ejemplo (Legacy):**
```
--- Menú Principal ---
1. Generar BAT desde EXE (modo Generador)
2. Ejecutar BAT existente (modo Ejecutar)
3. Salir
Elige una opción (1-3): 1
Modo Generador: Selecciona el archivo .exe...
Ingresa la ruta completa del archivo .exe (ej: C:\Path\To\miapp.exe): C:\Windows\notepad.exe
Generando BAT...
BAT creado en: C:\Windows\notepad_RunAsInvoker.bat
[MessageBox: ¿Deseas ejecutar el BAT ahora?]
```

---

<a id="troubleshooting"></a>
<img src="https://readme-typing-svg.herokuapp.com?font=Lexend+Giga&size=25&pause=1000&color=CCA9DD&vCenter=true&width=435&height=25&lines=TROUBLESHOOTING" width="450"/>

---

- **"bcryptprimitives.dll not found" (Legacy)**: Usa Go 1.20.14.
- **Diálogos no abren (Legacy)**: Ingresa ruta manual.
- **Compilación con x/sys (Legacy)**: `go get golang.org/x/sys@v0.7.0`.
- **Errores gráficos en Main (Win7/8 experimental)**: Habilita OpenGL o usa `-tags no_opengl`; soporte limitado.
- **MinGW no encontrado (Main)**: Instala paquetes listados.
- **Panic en runtime**: Verifica struct `OPENFILENAME` (Legacy) o dependencias Fyne (Main).
- **Pruebas**: Genera BAT de `notepad.exe` y verifica sin UAC.

---

## Licencia

MIT License.

## Contribuciones

Reporta issues en [GitHub](https://github.com/s7lver2/WinPass/issues). Pull requests bienvenidos.

¡Gracias por usar WinPass! Si hay bugs, abre una Issue. 🚀
