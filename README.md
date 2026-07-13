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

- ![operation](https://github.com/theextendedname/arcDPS-Installer/blob/main/arcDPS-Installer-Demo.gif)
Enjoy!
