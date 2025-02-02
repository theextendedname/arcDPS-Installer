package main

import (
		"io"
		"io/ioutil"
		"net/http"
		"net/url"
		"os"
		"os/exec"
		"path/filepath"
		"fmt"
		"golang.org/x/sys/windows/registry"
		"bufio"
		"strings"
		"strconv"
		"log"
		"encoding/json"
        

)
var version string // Declare version variable

type Config struct {// Configuration structure
        GW2PathOverRide string `json:""`        
}

func getInstallPath() (string, error) {
	//read a key from the windows registry
        key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\ArenaNet\Guild Wars 2`, registry.QUERY_VALUE)
        if err != nil {
                return "", fmt.Errorf("error opening registry key: %w", err)
        }
        defer key.Close()

        path, _, err := key.GetStringValue("Path")
        if err != nil {
                return "", fmt.Errorf("error reading registry value: %w", err)
        }

        return path, nil
}

func loadConfig(filepath string) (*Config, error) {
	//load config from the arcDPS-Installer-Config.json
	// Check if the config file exists.  If not, return a default config.
    if _, err := os.Stat(filepath); os.IsNotExist(err) {        
        return &Config{
            GW2PathOverRide: "",
           
        }, nil // Return default config
    }

        data, err := ioutil.ReadFile(filepath)
        if err != nil {
                return nil, fmt.Errorf("error reading config file: %w", err)
        }

        var config Config
        err = json.Unmarshal(data, &config)
        if err != nil {
                return nil, fmt.Errorf("error unmarshalling config: %w", err)
        }

        return &config, nil
}

func saveConfig(config *Config, filepath string) error {
        data, err := json.MarshalIndent(config, "", "    ") // Use indent for pretty formatting
        if err != nil {
                return fmt.Errorf("error marshalling config: %w", err)
        }
		
        err = ioutil.WriteFile(filepath, data, 0644) // 0644 permissions: rw-r--r--
        if err != nil {
                return fmt.Errorf("error writing config file: %w", err)
        }

        return nil
}


func folderPickerWindows() (string, error) {
   // ... (Not sure using powershell winform is the best way to do this but it seems to work)

    psScript := `
        Add-Type -AssemblyName System.Windows.Forms
        $dialog = New-Object System.Windows.Forms.FolderBrowserDialog
        $dialog.Description = "Select a folder"
        $result = $dialog.ShowDialog()
        if ($result -eq "OK") {
            $dialog.SelectedPath
        }
    `
	//run the powershell command
    cmd := exec.Command("powershell", "-Command", psScript)

    output, err := cmd.CombinedOutput()
    if err != nil {
        return "", fmt.Errorf("folder picker failed: %v\nOutput: %s", err, output)
    }

    path := strings.TrimSpace(string(output))

    if path == "" { // User cancelled. Handle this in main		
        return "", nil
    }

    // Path Validation (Still Essential)
    if _, err := os.Stat(path); os.IsNotExist(err) {
        return "", fmt.Errorf("selected path does not exist: %s", path)
    } else if err != nil {
        return "", fmt.Errorf("error checking path: %w", err)
    } else {
        fileInfo, err := os.Stat(path)
        if err != nil {
            return "", fmt.Errorf("error getting file info: %w", err)
        }
        if !fileInfo.IsDir() {
            return "", fmt.Errorf("selected path is not a directory: %s", path)
        }
    }

    return path, nil
}



func getResponseURI(url string) (string, error) {
        resp, err := http.Get(url)
        if err != nil {
                return "", fmt.Errorf("error during GET request: %w", err)
        }
        defer resp.Body.Close() // Important: Close the response body

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("unexpected status code: %d %s", resp.StatusCode, resp.Status)
    }

        return resp.Request.URL.String(), nil // Return the final URL (after redirects)
}

func directoryExists(path string) bool {
        _, err := os.Stat(path)
        if err == nil {
                return true // The directory exists.
        }
        if os.IsNotExist(err) {
                return false // The directory does NOT exist.
        }
        // Some other error occurred (e.g., permissions issue).
        return false // Treat as not exists for simplicity.  You might want to handle other errors differently in your specific use case.
}

func downloadFile(dlFilepath string, url string) error {

        // Get the data
        resp, err := http.Get(url)
        if err != nil {
                return err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return fmt.Errorf("bad status: %s", resp.Status)
        }


        // Check if game directory exixts
		dir2Test := filepath.Dir(dlFilepath)
		 if directoryExists(dir2Test) {
                fmt.Println("Directory exists:", dir2Test)
        } else {
                fmt.Println("Directory does NOT exist:", dir2Test)
        }

        out, err := os.Create(dlFilepath)
        if err != nil {
                return err
        }
        defer out.Close()

        // Write the body to file
        _, err = io.Copy(out, resp.Body)
        if err != nil {
                return err
        }

        return nil
}
func install_arcDPS(installDir string, urlString string) error {
	var fileDestinationPath string = ""
	 // Extract the filename from the URL  
    parsedURL, err := url.Parse(urlString)
    if err == nil { // Handle potential URL parsing errors
        fileName := filepath.Base(parsedURL.Path)		
        fileDestinationPath = filepath.Join(installDir , fileName) // Join path components correctly
		fmt.Println("Final URL:", urlString)
		fmt.Println("arcDPS Dll destination: ", fileDestinationPath)
    }

	err = downloadFile(fileDestinationPath, urlString)
	
	if err != nil {
		return err
	}
	return nil
}
func install_Healing_addon(installDir string, urlString string) error {
	var fileDestinationPath string = ""
	var Addon_Version string = ""
	
	versionURL, err := getResponseURI(urlString)
        if err != nil {
                fmt.Println("Error:", err)
        } else {
                fmt.Println("Final URL:", versionURL)
        }
		
		// Extract the version from the URL  
    parsedVersionURL, err := url.Parse(versionURL)
    if err == nil { // Handle potential URL parsing errors
        Addon_Version = filepath.Base(parsedVersionURL.Path)
		//set the new url to latest version
		urlString = "https://github.com/Krappa322/arcdps_healing_stats/releases/download/" + Addon_Version + "/arcdps_healing_stats.dll"		
        fmt.Println("Healing Add-on Version: ", Addon_Version)
    }
	 // Extract the filename from the URL  
    parsedURL, err := url.Parse(urlString)
    if err == nil { // Handle potential URL parsing errors
        fileName := filepath.Base(parsedURL.Path)		
        fileDestinationPath = filepath.Join(installDir , fileName) // Join path components correctly
		fmt.Println("Healing Add-on Dll destination: ", fileDestinationPath)
    }

	
	err = downloadFile(fileDestinationPath, urlString)
	
	if err != nil {
		return err
	}
	return nil
}


func insatll_BoonTable_Addon(installDir string, urlString string) error {
	var fileDestinationPath string = ""
	var Addon_Version string = ""
	
	versionURL, err := getResponseURI(urlString)
        if err != nil {
                fmt.Println("Error:", err)
        } else {
                fmt.Println("Final URL:", versionURL)
        }
		
		// Extract the version from the URL  
    parsedVersionURL, err := url.Parse(versionURL)
    if err == nil { // Handle potential URL parsing errors
        Addon_Version = filepath.Base(parsedVersionURL.Path)
		//set the new url to latest version
		urlString = "https://github.com/knoxfighter/GW2-ArcDPS-Boon-Table/releases/download/" + Addon_Version + "/d3d9_arcdps_table.dll"		
        fmt.Println("BoonTable Add-on Version: ", Addon_Version)
    }
	 // Extract the filename from the URL  
    parsedURL, err := url.Parse(urlString)
    if err == nil { // Handle potential URL parsing errors
        fileName := filepath.Base(parsedURL.Path)		
        fileDestinationPath = filepath.Join(installDir , fileName) // Join path components correctly
		fmt.Println("BoonTable Add-on Dll destination: ", fileDestinationPath)
    }
	
	err = downloadFile(fileDestinationPath, urlString)
	
	if err != nil {
		return err
	}
	return nil
}
func getUserChoice(prompt string, defaultChoice int) int {
        reader := bufio.NewReader(os.Stdin)

        for {
                fmt.Print(prompt)
                input, _ := reader.ReadString('\n')
                input = strings.TrimSpace(input)

                if input == "" {
                        return defaultChoice
                }

                choice, err := strconv.Atoi(input)
                if err != nil || choice < 1 || choice > 7 {
                        fmt.Println("Invalid input. Please enter a number between 1 and 7.")
                        continue // Ask again
                }

                return choice
        }
}
func clearScreenANSI() {
    fmt.Print("\033[H\033[2J") // Clear screen and move cursor to top-left
}

func updatePromptString(installDir string) string{
	//declare prompt	
prompt := `1) arcDPS Add-on only
2) Healing Add-on only
3) Boon-Table Add-on only
4) Remove All Add-ons` + "\n"
prompt += `5) Change GW2 install Path: [` + installDir + `]` + "\n"
prompt +=  `6) Exit or Ctl+C 
7) Install/Update All
Choose a mode  1 - 7 (7 is default):`

	return prompt 
}

func updateInstallDir() string{
	installDir, err := folderPickerWindows()
		if err != nil {
				log.Fatal(err)
		}

		if installDir == "" {
				fmt.Println("No folder selected.")
		} else {
				fmt.Println("Selected folder:", installDir)
		}
	return installDir
}

func main() {
	var installDir string = ""
	
	var arcDPS_urlString string= "https://www.deltaconnected.com/arcdps/x64/d3d11.dll"
	var HealingAddon_urlString string= "https://github.com/Krappa322/arcdps_healing_stats/releases/latest"
	var BoonTableAddon_urlString string= "https://github.com/knoxfighter/GW2-ArcDPS-Boon-Table/releases/latest"
	
	//returns the absolute path to the executable file itself
	exePath, err := os.Executable()
	if err != nil {
			log.Fatal("Error getting executable path:", err)
	}
	//get the directory containing the executable
	exeDir := filepath.Dir(exePath)
	configFilePath := exeDir +"/arcDPS-Installer-Config.json" // config path
	//load config 
	config, err := loadConfig(configFilePath)
	if err != nil {
			fmt.Println("Error loading config:", err)
		
	}else {
		if config.GW2PathOverRide == "" {
			// no config found use what's in the registry
			fmt.Println("Config file not found. Reading from registry.")
			installPath, err := getInstallPath()
			if err != nil {
					fmt.Println("Error:", err)
					//install directory not found. Warn user
					installDir = "Tragedy! Your GW2 install directory is missing. Please use Opt 5"				
			} else {
				 installDir = filepath.Dir(installPath)
				 fmt.Println("Guild Wars 2 Install Path:", installDir)
			}
		}else{
		//config found use it
		installDir = config.GW2PathOverRide
		}
	}
	 
		
	//PrintHeader
fmt.Println("arcDPS-Instaler Version " , version) 
fmt.Println("by Extended")
fmt.Println("This app can install, update and remove arcDPS and some Add-ons ")
fmt.Println("This app has no affiliation with the arcDPS project or it's Add-ons")
fmt.Println("********************************************************************")
 for { 
//declare prompt	
//prompt := `1) arcDPS Add-on only
//2) Healing Add-on only
//3) Boon-Table Add-on only
//4) Remove All Add-ons
//5) Exit or Ctl+C 
//6) Install/Update All
//Choose a mode  1, 2, 3, 4, 5, or 6 (6 is default):`
prompt := updatePromptString(installDir)
				
		choice := getUserChoice(prompt, 7)
		switch choice {
			case 1:
					err = install_arcDPS(installDir, arcDPS_urlString)
					if err != nil {
							fmt.Println("Error downloading arcDPS:", err)
					} else {
							fmt.Println("arcDPS downloaded")
					}
					fmt.Println("--------------------------------------------------")
			case 2:
					err = install_Healing_addon(installDir, HealingAddon_urlString)
					if err != nil {
							fmt.Println("Error downloading Healing Add-on:", err)
					} else {
							fmt.Println("Healing Add-on downloaded")
					}
					fmt.Println("--------------------------------------------------")
			case 3: 
					   err = insatll_BoonTable_Addon(installDir, BoonTableAddon_urlString)
						if err != nil {
								fmt.Println("Error downloading BoonTable Add-on:", err)
						} else {
								fmt.Println("BoonTable Add-on downloaded")
						}
						fmt.Println("--------------------------------------------------")   
			 case 4: //remove All
			 
					files := []string{installDir + "\\d3d11.dll", installDir + "\\arcdps_healing_stats.dll", installDir + "\\d3d9_arcdps_table.dll", installDir + "\\bin64\\d3d11.dll", installDir + "\\bin64\\arcdps_healing_stats.dll", installDir + "\\bin64\\d3d9_arcdps_table.dll"}
						for _, file := range files {
								err := os.Remove(file)
								if err != nil {
										if os.IsNotExist(err) {
												fmt.Printf("File %s does not exist, skipping.\n", file)
										} else if os.IsPermission(err) {
												log.Fatalf("Permission denied deleting %s: %v", file, err)
										} else {
												log.Fatalf("Error deleting %s: %v", file, err)
										}
								} else {
										fmt.Printf("File %s deleted successfully.\n", file)
								}
						}
			 case 5: //changes Add-on install path
					installDirTemp := installDir //save value incase urser cancels filepickre
					installDir = updateInstallDir()
					if installDir == "" {
						//user canclled dialog. reset installDir
						installDir = installDirTemp
						clearScreenANSI()
					} else {
						//update config var and file
						config.GW2PathOverRide = installDir
						err = saveConfig(config, configFilePath)
						clearScreenANSI()
						if err != nil {
								fmt.Println("Error saving config:", err)
						} else {								
								fmt.Println("Config saved successfully.")
						}
					}
					
			 case 6: //Exit
					fmt.Println("Good Bye....")
					os.Exit(0) // Exit the the program
			default:
					//install all
					err = install_arcDPS(installDir, arcDPS_urlString)
					if err != nil {
							fmt.Println("Error downloading arcDPS:", err)
					} else {
							fmt.Println("arcDPS downloaded")
					}
					fmt.Println("--------------------------------------------------")
					
				   err = install_Healing_addon(installDir, HealingAddon_urlString)
					if err != nil {
							fmt.Println("Error downloading Healing Add-on:", err)
					} else {
							fmt.Println("Healing Add-on downloaded")
					}
					fmt.Println("--------------------------------------------------")
					
					   err = insatll_BoonTable_Addon(installDir, BoonTableAddon_urlString)
					if err != nil {
							fmt.Println("Error downloading BoonTable Add-on:", err)
					} else {
							fmt.Println("BoonTable Add-on downloaded")
					}
					fmt.Println("--------------------------------------------------")
		}
		fmt.Println("********************************************************************")
		fmt.Println("********************************************************************")	
		fmt.Println("Action Complete....")
	}	
	
       
		
		
		
		
	
	
   
	

}