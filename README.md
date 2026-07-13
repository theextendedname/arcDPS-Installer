# GW2 Addon Downloader

A simple app to download the following Guild Wars 2 add-ons and place them in your GW2 directory.

*   **arcDPS:** [https://www.deltaconnected.com/arcdps/x64](https://www.deltaconnected.com/arcdps/x64)
*   **Healing Add-on:** [https://github.com/Krappa322/arcdps_healing_stats](https://github.com/Krappa322/arcdps_healing_stats)
*   **Boon-Table Add-on:** [https://github.com/knoxfighter/GW2-ArcDPS-Boon-Table](https://github.com/knoxfighter/GW2-ArcDPS-Boon-Table)

**[Download Latest Release](https://github.com/theextendedname/arcDPS-Installer/releases/latest)**

- Download the app
- Run the app
- Press Enter to Install or Update Addons
- Use ↑↓ Arrow or JK to select options
- Press Ctrl + Z to switch between Main Menu and Status Views 
- Press Q to Quit

## Linux

Guild Wars 2 is detected at `~/.local/share/Steam/steamapps/common/Guild Wars 2/` by default. If it is not there, the installer opens a GUI folder picker using `zenity` or `kdialog`.

Build the Linux binary with:

```bash
GOOS=linux GOARCH=amd64 go build -o arcDPS-Installer-linux .
```
Linux handles this differently from Windows.
    
    On Windows, a console-subsystem executable
    automatically receives a console window. Linux desktop
    environments generally do not create a terminal when
    an executable is double-clicked.
    
    The minimal Linux solution is a .desktop launcher
    with Terminal=true:
    
    ini
    [Desktop Entry]
    Type=Application
    Name=arcDPS Installer
    Comment=Install and update Guild Wars 2 add-ons
    Exec=/full/path/to/arcDPS-Installer-linux
    Icon=utilities-terminal
    Terminal=true
    Categories=Utility;
    
    
    When launched from the application menu or desktop,
    Linux opens the utility inside the user’s configured
    terminal.
    
- ![operation](https://github.com/theextendedname/arcDPS-Installer/blob/main/arcDPS-Installer-Demo.gif)
Enjoy!
