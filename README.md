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

Este proyecto es una aplicación GUI en Go para Windows que genera archivos `.bat` con compatibilidad `RunAsInvoker` para ejecutar ejecutables (`.exe`) sin elevación de privilegios UAC. Es compatible con gran parte de los sitemas operativos Windows, entre ellos **windows 7, windows 10, y windows 11** (requesitos mínimos: Windows 7 KB2533623).

> [!IMPORTANT]
> No me hago responsable de los usos que se le puedan dar a esta herramienta, y solo apruebo su uso con regulación o sentido común
---

<a id="versions"></a>
<img src="https://readme-typing-svg.herokuapp.com?font=Lexend+Giga&size=25&pause=1000&color=CCA9DD&vCenter=true&width=435&height=25&lines=VERSIONS" width="450"/>

---

Este repositorio incluye dos versiones del proyecto, cada una optimizada para diferentes entornos y compatibilidades. Recomendamos personalmente el uso de la version **Main**

1. **Main (directorio padre: `../`)**: Versión GUI con Fyne (interfaz gráfica moderna con pestañas). Soporta **Windows 10/11** de forma estable. Soporte experimental para **Windows 8/7**, pero con requesito de aceleracion gráfica (OpenGL/DirectX). Ideal para usuarios que prefieren una interfaz visual intuitiva.

2. **Legacy (directorio actual: `Legacy/`)**: Versión CLI secundaria con diálogos nativos de Windows y fallback manual. Soporta **Windows 7 SP1+** de forma nativa. No requiere MinGW. Útil para sistemas legacy o entornos sin GUI.


Para navegar:
```shell
cd ..      # Para Main (principal, Win10/11)
cd Legacy  # Para Legacy (secundaria, Win7)
```

> [!NOTE]
> La versión Main usa su propio `go.mod` independiente. Compila por separado para evitar conflictos de dependencias.

---

<a id="compilation"></a>
<img src="https://readme-typing-svg.herokuapp.com?font=Lexend+Giga&size=25&pause=1000&color=CCA9DD&vCenter=true&width=435&height=25&lines=COMPILATION" width="450"/>

---

> [!IMPORTANT]
> Compila con Go 1.20.14 para soporte en Windows 7; versiones posteriores fallan por `bcryptprimitives.dll`.

> [!CAUTION]
> Para compilar la legacy version, necesitas Go 1.20.14 y `golang.org/x/sys@v0.7.0`.

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
