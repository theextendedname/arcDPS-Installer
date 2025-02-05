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
 // Define color codes (you can customize these)
const (
		Reset  = "\033[0m"
		Red    = "\033[31m"
		Green  = "\033[32m"
		Yellow = "\033[33m"
		Blue   = "\033[34m"
)
//define urls 
const (
arcDPS_urlString string = "https://www.deltaconnected.com/arcdps/x64/d3d11.dll"
HealingAddon_urlString string = "https://github.com/Krappa322/arcdps_healing_stats/releases/latest"
BoonTableAddon_urlString string = "https://github.com/knoxfighter/GW2-ArcDPS-Boon-Table/releases/latest"
arcDPSInstaller_urlString string = "https://github.com/theextendedname/arcDPS-Installer/releases/latest"
)		
var arcDPS_DlUrlAry = []string{"","https://www.deltaconnected.com/arcdps/x64"}
var HealingAddon_DlUrlAry = []string{"https://github.com/Krappa322/arcdps_healing_stats/releases/download/","/arcdps_healing_stats.dll"}
var BoonTableAddon_DlUrlAry = []string{"https://github.com/knoxfighter/GW2-ArcDPS-Boon-Table/releases/download/","/d3d9_arcdps_table.dll"}
var arcDPSInstaller_DlUrlAry =  []string{"",""}


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

func versionToInt(versionStr string) int {
	//used to compair app version numbers
	if versionStr == "" {
		//set minimum version		
		versionStr = "0.0.1"		
	}
    parts := strings.Split(versionStr, ".")
    var paddedParts []string
    for _, part := range parts {
        paddedParts = append(paddedParts, fmt.Sprintf("%03s", part))
    }
    concatenated := strings.Join(paddedParts, "")
    intVersion, _ := strconv.Atoi(concatenated)
    return intVersion

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
                return fmt.Errorf(Red + "bad status: %s" + Reset, resp.Status)
        }


        // Check if game directory exixts
		dir2Test := filepath.Dir(dlFilepath)
		 if directoryExists(dir2Test) {
                fmt.Println("Directory exists:", dir2Test)
        } else {
                fmt.Println(Red + "Directory does NOT exist:" + Reset, dir2Test)
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

func getLatestAppVer_Github(urlString string) (string, error){
	//urlString := "https://github.com/theextendedname/arcDPS-Installer/releases/latest"
	var Addon_Version string = ""
	
	versionURL, err := getResponseURI(urlString)
        if err != nil {
                
				return "", fmt.Errorf(Red + "Update Check Error:" + Reset, err)
        } 
		
		// Extract the version from the URL  
    parsedVersionURL, err := url.Parse(versionURL)
    if err == nil { // Handle potential URL parsing errors
	
        Addon_Version = filepath.Base(parsedVersionURL.Path)
		return Addon_Version, nil
		//fmt.Println(Addon_Version)
	 }	
	//return minimum version
	return "v0.0.1", err
   
	
}

func insatll_Addon(installDir string, urlString string, addonName string, DlUrlAry []string ) error {
	var fileDestinationPath string = ""		
	if DlUrlAry[0] != "" {	//add-on is on github	
		Addon_Version , err:= getLatestAppVer_Github(urlString)
		if err != nil {
			return fmt.Errorf(Red + "Error checking " + addonName + " Add-on version:" + Reset, err)
		}else { // Handle potential URL parsing errors		
		//set the new url to latest version
		urlString = DlUrlAry[0] + Addon_Version + DlUrlAry[1]		
		fmt.Println(addonName, "Add-on Version:", Addon_Version)
		fmt.Println("Final URL:", urlString)
		}
	}
	 // Extract the filename from the URL  
    parsedURL, err := url.Parse(urlString)
    if err == nil { // Handle potential URL parsing errors
        fileName := filepath.Base(parsedURL.Path)		
        fileDestinationPath = filepath.Join(installDir , fileName) // Join path components correctly
		if DlUrlAry[0] == ""{	//add-on is not on github
			//no version numb 
			fmt.Println("Final URL:", DlUrlAry[1])			
		} 
		fmt.Println(addonName, "Add-on Dll destination: [" + fileDestinationPath +"]")
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
				if strings.ToUpper(input) == "Q" {
                        return 6
                }

                choice, err := strconv.Atoi(input)
                if err != nil || choice < 1 || choice > 7 {
					    clearScreenANSI()
						PrintHeader()
                        fmt.Println(Yellow + "Invalid input. Please enter a number between 1 and 7." + Reset)
                        continue // Ask again
                }

                return choice
        }
}
func clearScreenANSI() {
    fmt.Print("\033[H\033[2J") // Clear screen and move cursor to top-left
}

func PrintHeader(){	
	if version == "" {
		//version is set at compile time so in could be empty during testing
		//versionInt will be set to 0.0.1
		version = "0.0.1"
	}
fmt.Println("arcDPS-Instaler Version " , version) 
headerStr := `by TheExtendedName 
Project website https://github.com/theextendedname/arcDPS-Installer
This app can install, update and remove arcDPS, Healing, and Boon-Table Add-ons 
This app has no affiliation with the arcDPS project or it's Add-ons
********************************************************************
********************************************************************`
fmt.Println(headerStr) 
}

func updatePromptString(installDir string) string{
	//declare prompt	
prompt := `1) arcDPS Add-on only
2) Healing Add-on only
3) Boon-Table Add-on only
4) Remove All Supported Add-ons` + "\n"
prompt += `5) Change GW2 install Path: [` + installDir + `]` + "\n"
prompt +=  `6) or Q to Quit app
7) Install/Update All` + "\n"
prompt += `Choose a mode  1 - 7 ` + Yellow + `(7 is default)` + `:` + Reset 

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
	//main go function
	var installDir string = "" //install path for GW2-64.exe
	
	PrintHeader()
	
	//returns the absolute path to the executable file itself
	exePath, err := os.Executable()
	if err != nil {
			log.Fatal("Error getting executable path:" , err)
	}
	//get the directory containing the executable
	exeDir := filepath.Dir(exePath)
	configFilePath := exeDir +"/arcDPS-Installer-Config.json" // config path
	//load config 
	config, err := loadConfig(configFilePath)
	if err != nil {
			fmt.Println(Red + "Error loading config:" + Reset, err)
		
	}else {
		if config.GW2PathOverRide == "" {
			// no config found use what's in the registry
			
			fmt.Println("Config file not found. Reading from registry.")
			installPath, err := getInstallPath()
			if err != nil || installPath == ""{
					fmt.Println(Red + "Registry Read Error:" + Reset, err)
					//install directory not found. Warn user
					installDir = Yellow + "Tragedy! Your GW2 install directory is missing. Please use Opt 5" + Reset				
			} else {
				
				 installDir = filepath.Dir(installPath)
				 fmt.Println("Guild Wars 2 Install Path:", installDir)
			}
		}else{
		//config found use it
		installDir = config.GW2PathOverRide
		}
	}
	 
	

 for { 
//declare prompt	
prompt := updatePromptString(installDir)
				
		choice := getUserChoice(prompt, 7)
		switch choice {
			case 1:
					fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
					//err = install_arcDPS(installDir, arcDPS_urlString)
					err = insatll_Addon(installDir, arcDPS_urlString, "arcDPS", arcDPS_DlUrlAry )
					if err != nil {
							fmt.Println(Red + "Error downloading arcDPS:" + Reset, err)
					} else {
							fmt.Println(Green + "arcDPS downloaded" + Reset)
					}
					fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
			case 2:
					fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
					//err = install_Healing_addon(installDir, HealingAddon_urlString)
					err = insatll_Addon(installDir, HealingAddon_urlString, "Healing", HealingAddon_DlUrlAry )
					if err != nil {
							fmt.Println(Red + "Error downloading Healing Add-on:" + Reset, err)
					} else {
							fmt.Println(Green + "Healing Add-on downloaded" + Reset)
					}
					fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
			case 3: 
					   fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
					   //err = insatll_BoonTable_Addon(installDir, BoonTableAddon_urlString)
					   err = insatll_Addon(installDir, BoonTableAddon_urlString, "Boon Table", BoonTableAddon_DlUrlAry )
						if err != nil {
								fmt.Println(Red + "Error downloading BoonTable Add-on:" + Reset, err)
						} else {
								fmt.Println(Green + "BoonTable Add-on downloaded" +Reset)
						}
						fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
			 case 4: //remove All
					fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
					files := []string{installDir + "\\d3d11.dll", installDir + "\\arcdps_healing_stats.dll", installDir + "\\d3d9_arcdps_table.dll", installDir + "\\bin64\\d3d11.dll", installDir + "\\bin64\\arcdps_healing_stats.dll", installDir + "\\bin64\\d3d9_arcdps_table.dll"}
						for _, file := range files {
								err := os.Remove(file)
								if err != nil {
										if os.IsNotExist(err) {
												fmt.Printf(Red + "File %s does not exist, skipping.\n" + Reset, file)
										} else if os.IsPermission(err) {
												log.Fatalf("Permission denied deleting %s: %v", file, err)
										} else {
												log.Fatalf("Error deleting %s: %v", file, err)
										}
								} else {
										fmt.Printf(Green + "File %s deleted successfully.\n" + Reset, file)
								}
						}
						fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
			 case 5: //changes Add-on install path
					installDirTemp := installDir //save value incase urser cancels filepicker
					installDir = updateInstallDir()
					if installDir == "" {
						//user canclled dialog. reset installDir
						installDir = installDirTemp
						clearScreenANSI()
						PrintHeader()
					} else {
						//update config var and file
						config.GW2PathOverRide = installDir
						err = saveConfig(config, configFilePath)
						clearScreenANSI()
						PrintHeader()
						if err != nil {
								fmt.Println(Red + "Error saving config:" + Reset, err)
						} else {								
								fmt.Println(Green + "Config saved successfully." + Reset)
						}
					}
					
			 case 6: //Exit
					fmt.Println(Green + "Good Bye...." + Reset)
					os.Exit(0) // Exit the the program
			default:
					
					//install all
					fmt.Println(Yellow + "--------------------------------------------------" + Reset)   
					
					//err = install_arcDPS(installDir, arcDPS_urlString)
					err = insatll_Addon(installDir, arcDPS_urlString, "arcDPS", arcDPS_DlUrlAry )
					if err != nil {
							fmt.Println(Red + "Error downloading arcDPS:" + Reset, err)
					} else {
							fmt.Println(Green + "arcDPS downloaded"  + Reset)
					}
					fmt.Println(Yellow + "--------------------------------------------------" + Reset)   					
				   
				   //err = install_Healing_addon(installDir, HealingAddon_urlString)
					err = insatll_Addon(installDir, HealingAddon_urlString, "Healing", HealingAddon_DlUrlAry )
					if err != nil {
							fmt.Println(Red + "Error downloading Healing Add-on:" + Reset, err)
					} else {
							fmt.Println(Green + "Healing Add-on downloaded" + Reset)
					}
					fmt.Println(Yellow + "--------------------------------------------------" + Reset) 
					
					   //err = insatll_BoonTable_Addon(installDir, BoonTableAddon_urlString)
					err = insatll_Addon(installDir, BoonTableAddon_urlString, "Boon Table", BoonTableAddon_DlUrlAry )
					if err != nil {
							fmt.Println(Red + "Error downloading BoonTable Add-on:" + Reset, err)
					} else {
							fmt.Println(Green + "BoonTable Add-on downloaded"  + Reset)
					}
					fmt.Println(Yellow + "--------------------------------------------------" + Reset) 
					//check for app updates
					
					latestAppVersion , err:= getLatestAppVer_Github(arcDPSInstaller_urlString)
					if err != nil {
							fmt.Println(Red + "Error checking arcDPS-Installer version:" + Reset, err)
					}
					//version strings to int 
					latestAppVersionInt := versionToInt(latestAppVersion[1:])
					versionInt := versionToInt(version)
					//version is set at compile time so in could be empty during testing
					//versionInt will be set to 0.0.1
					if versionInt < latestAppVersionInt {
						fmt.Println(Yellow + "New Version " + latestAppVersion + " Avalible @  https://github.com/theextendedname/arcDPS-Installer/releases/latest" + Reset) 
					}	
		}
				
		fmt.Println(Blue + "Action Complete...." + Reset)
	}	
	
       
		
		
		
		
	
	
   
	

}