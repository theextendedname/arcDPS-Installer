package main

import (
        "io"
		"io/ioutil"		
		"net/http"
		"net/url"
		"os"				
		"path/filepath"
		"fmt"
		"time"
		"log"
		"strings"
		"strconv"
		"encoding/json"		
        tea "github.com/charmbracelet/bubbletea"
		"github.com/charmbracelet/lipgloss"
		"golang.org/x/sys/windows/registry"	
		"arcDPS_Installer/folderpicker"			
)

var version string // Declare version variable

type Config struct {// Configuration structure
        GW2PathOverRide string `json:""`        
}

//var installDir string = "D:\\Games\\Guild Wars 2" //install path for GW2-64.exe
var installDir string = "C:\\Program Files\\Guild Wars 2" //default install path for GW2-64.exe
var configFilePath string = ""

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
var infoCache string  = ""			//cache for info view		

type model struct {
        choices   []string
        cursor    int
        selected  map[int]struct{} 	// This is now unnecessary, we use the 'choice' variable instead
        choice    string           	// The currently selected choice
		status 	string				//the progres of downloads		
		altscreen  bool
		quitting   bool
		err    error
}
type statusMsg string 
type errMsg struct{ err error }


// For messages that contain errors 
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

var (
	keywordStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204")).Background(lipgloss.Color("235"))
	helpStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	selestedStyle = lipgloss.NewStyle().Bold(true).Background(lipgloss.Color("55"))
	sucessStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("104"))
	warrningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("209"))
	errorStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196")).Background(lipgloss.Color("219"))
	dlSeperatorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
)

func getResponseURI(url string) (string, error) {
        resp, err := http.Get(url)
        if err != nil {
                return "", fmt.Errorf(errorStyle.Render("error during GET request: %w"), err)
        }
        defer resp.Body.Close() // Important: Close the response body

    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf(errorStyle.Render("unexpected status code: %d %s"), resp.StatusCode, resp.Status)
    }

        return resp.Request.URL.String(), nil // Return the final URL (after redirects)
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
		//only return the game dir
		path  = filepath.Dir(path)
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

func downloadFile(dlFilepath string, url string) (string, error) {
		var infoStr string = ""
        // Get the data
		c := &http.Client{
			Timeout: 120 * time.Second,
		}
        resp, err := c.Get(url)
        if err != nil {
                return infoStr, err
        }
        defer resp.Body.Close()

        if resp.StatusCode != http.StatusOK {
                return infoStr , fmt.Errorf(errorStyle.Render("bad status: %s"), resp.Status)
        }


        // Check if game directory exixts
		dir2Test := filepath.Dir(dlFilepath)
		 if directoryExists(dir2Test) {
               infoStr += "Directory exists: " + dir2Test + "\n"
        } else {
               return infoStr , fmt.Errorf(errorStyle.Render("Directory does NOT exist: %s"), dir2Test)
        }

        out, err := os.Create(dlFilepath)
        if err != nil {
                return infoStr , err
        }
        defer out.Close()

        // Write the body to file
        _, err = io.Copy(out, resp.Body)
        if err != nil {
                return infoStr , err
        }

        return infoStr, nil
}

func getLatestAppVer_Github(urlString string) (string, error){
	//urlString := "https://github.com/theextendedname/arcDPS-Installer/releases/latest"
	var Addon_Version string = ""
	
	versionURL, err := getResponseURI(urlString)
        if err != nil {
                
				return "", fmt.Errorf(errorStyle.Render( "Update Check Error: %s"), err)
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

func updateInstallPath(config *Config , configFilePath string) tea.Cmd {
	//changes Add-on install path
	// installDir global var
	return func() tea.Msg {

		var infoStr string = ""		
		//launch folderPickerWindows
		//installDirTemp, err := folderPickerWindows()	
		installDirTemp, err := folderpicker.Prompt("Select Your GW2 Game folder")
			if err != nil {
					//log.Println(err)
					return errMsg{fmt.Errorf(errorStyle.Render("File Picker Error. Press Ctrl-Z : %s"), err)}
			}

			if installDirTemp != "" {			
				//fmt.Println("Selected folder:", installDir)
				infoStr += warrningStyle.Render("Selected folder: " + installDirTemp ) + "\n"	
			}
		if installDirTemp != "" {
			//only empty if user canclled dialog. no change to installDir		
			//set new installDir
			installDir = installDirTemp
			//update config var and file
			config.GW2PathOverRide = installDir
			err = saveConfig(config, configFilePath)		
			if err != nil {
					
					//log.Println("Error saving config: %v" ,configFilePath, err)				
					return errMsg{fmt.Errorf(errorStyle.Render("Error saving config: %v %s"),configFilePath, err)}
			} else {								
					//fmt.Println(Green + "Config saved successfully." + Reset)
					infoStr += sucessStyle.Render("Config saved successfully." + configFilePath) + "\n"
					return statusMsg(infoStr)				
			}
		}
		return statusMsg("/n")	
	}				
}

func checkAppUpdates(urlString string ) tea.Cmd {
	//tea.cmd wraper called during install/update all
	//arcDPSInstaller_urlString 
	return func() tea.Msg {
		var infoStr string = ""
		latestAppVersion , err:= getLatestAppVer_Github(urlString)
		if err != nil {// Handle potential URL parsing errors	
			//return infoStr, fmt.Errorf(errorStyle.Render("Error checking " + addonName + " Add-on version: %s"), err)
			return errMsg{fmt.Errorf(errorStyle.Render("Error checking app updates: %s"), err)}
		}else { 	
				latestAppVersionInt := versionToInt(latestAppVersion[1:])
				versionInt := versionToInt(version)
				//version is set at compile time so in could be empty during testing
				//versionInt will be set to 0.0.1
				if versionInt < latestAppVersionInt {
					//fmt.Println(Yellow + "New Version " + latestAppVersion + " Avalible @  https://github.com/theextendedname/arcDPS-Installer/releases/latest" + Reset) 
					infoStr += warrningStyle.Render("New Version " + latestAppVersion + " Avalible @  https://github.com/theextendedname/arcDPS-Installer/releases/latest") + "\n"					
					infoCache += infoStr
					return statusMsg(infoStr)	
				}	
		}
		return statusMsg("/n")	
	}

}
func insatll_Addon(installDir string, urlString string, addonName string, DlUrlAry []string )(string, error) {
	var fileDestinationPath string = ""	
	var infoStr string = ""	
	if DlUrlAry[0] != "" {	//add-on is on github	
		Addon_Version , err:= getLatestAppVer_Github(urlString)
		if err != nil {
			return infoStr, fmt.Errorf(errorStyle.Render("Error checking " + addonName + " Add-on version: %s"), err)
		}else { // Handle potential URL parsing errors		
		//set the new url to latest version
		urlString = DlUrlAry[0] + Addon_Version + DlUrlAry[1]		
		infoStr += addonName + " Add-on Version: " + Addon_Version + "\n"
		infoStr += "Final URL: " + urlString + "\n"
		}
	}
	 // Extract the filename from the URL  
    parsedURL, err := url.Parse(urlString)
    if err != nil { // Handle potential URL parsing errors
		return infoStr, err        
    }else{
		fileName := filepath.Base(parsedURL.Path)		
        fileDestinationPath = filepath.Join(installDir , fileName) // Join path components correctly
		if DlUrlAry[0] == ""{	//add-on is not on github
			//no version numb 
			infoStr += "Final URL: " + DlUrlAry[1]  + "\n"		
		} 
		infoStr += addonName + " Add-on Dll destination: [" + fileDestinationPath +"]" + "\n"
	}
	
	dlInfo, err := downloadFile(fileDestinationPath, urlString)
	infoStr += dlInfo 
	if err != nil {		
		return infoStr, err
	}
	return infoStr, nil
}
func removeAddOns(installDir string) string {
	var infoStr string	= ""							
	files := []string{installDir + "\\d3d11.dll", installDir + "\\arcdps_healing_stats.dll", installDir + "\\d3d9_arcdps_table.dll", installDir + "\\bin64\\d3d11.dll", installDir + "\\bin64\\arcdps_healing_stats.dll", installDir + "\\bin64\\d3d9_arcdps_table.dll"  , filepath.Dir(installDir) + "\\d3d11.dll", filepath.Dir(installDir) + "\\arcdps_healing_stats.dll", filepath.Dir(installDir) + "\\d3d9_arcdps_table.dll"}
	for _, file := range files {
		//infoStr += filepath.Base(filepath.Dir(file))
			if strings.Contains(file , "\\bin64\\bin64" ) {continue}   
			
			if filepath.Base(filepath.Dir(file)) == "Guild Wars 2" || filepath.Base(filepath.Dir(file)) == "bin64" {
			
				err := os.Remove(file)
				if err != nil {
						if os.IsNotExist(err) {
								//fmt.Printf(Red + "File %s does not exist, skipping.\n" + Reset, file)
								//log.Println("File %s does not exist, skipping. : %v \n", file, err)												
								infoStr += errorStyle.Render(file) + errorStyle.Render(" does not exist, skipping.") + "\n"
						} else if os.IsPermission(err) {
								//log.Println("Permission denied deleting %s: %v", file, err)
								infoStr += errorStyle.Render("Permission denied deleting ") + errorStyle.Render(file) + "\n"
						} else {
								//log.Println("Error deleting %s: %v", file, err)
								infoStr += errorStyle.Render("Error deleting ") + errorStyle.Render(file)  + "\n"
						}
				} else {
						infoStr += sucessStyle.Render(file) + sucessStyle.Render(" deleted successfully.")  + "\n"
						
				}
			}
	}	
	
	infoStr += dlSeperatorStyle.Render("---------------------------------------------------") + "\n"
	infoStr += "\n" + sucessStyle.Render("Add-On Removal Complete.")  + "\n"
	return infoStr
}

func downloadFunc(opt int ) tea.Cmd {
	return func() tea.Msg {
		
       switch opt {
		case 0:
				infoStr := sucessStyle.Render( "Install/Update All Add-Ons") + "\n"
				infoCache += infoStr				
				return statusMsg(infoStr)
		case 1:
				infoStr, err := insatll_Addon(installDir, arcDPS_urlString, "arcDPS", arcDPS_DlUrlAry )
				if err != nil {		
						return errMsg{fmt.Errorf(errorStyle.Render("Error downloading arcDPS: %s" ), err)}
						//return statusMsg(infoStr) , fmt.Errorf(Red + "Error downloading arcDPS:" + Reset, err)
				} else {
						infoStr += sucessStyle.Render("arcDPS downloaded") + "\n"
						infoStr += dlSeperatorStyle.Render("---------------------------------------------------") + "\n"	
						infoCache += infoStr
						return statusMsg( infoStr)
				}
		case 2:
				infoStr, err := insatll_Addon(installDir, HealingAddon_urlString, "Healing", HealingAddon_DlUrlAry )
				if err != nil {
						 return errMsg{fmt.Errorf(errorStyle.Render("Error downloading Healing Add-on: %s"), err)}
						//return statusMsg(infoStr) ,fmt.Errorf(Red + "Error downloading Healing Add-on:" + Reset, err)
				} else {
						infoStr += sucessStyle.Render( "Healing Add-on downloaded") + "\n"
						infoStr += dlSeperatorStyle.Render("---------------------------------------------------") + "\n"	
						infoCache += infoStr						
						return statusMsg(infoStr)
				}
		case 3:
				infoStr, err := insatll_Addon(installDir, BoonTableAddon_urlString, "Boon Table", BoonTableAddon_DlUrlAry )
				if err != nil {
						 return errMsg{fmt.Errorf(errorStyle.Render("Error downloading BoonTable Add-on: %s" ), err)}
						//return statusMsg(infoStr) , fmt.Errorf(Red + "Error downloading BoonTable Add-on:" + Reset, err)
				} else {
					
					infoStr += sucessStyle.Render("BoonTable Add-on downloaded") + "\n"
					infoStr += dlSeperatorStyle.Render("---------------------------------------------------") + "\n"		
					infoCache += infoStr
					return statusMsg(infoStr)
				}
		case 4:
				//infoStr :=sucessStyle.Render( "delete addons") + "\n"
				infoStr := removeAddOns(installDir)
				return statusMsg(infoStr)
	
		}
        return statusMsg("No Process")
    }
}

func initialModel() model {
	
	return model{
		choices:  []string{"Install/Update All", "Install arcDPS Add-on", "Install Healing Add-on", "Install Boon-Table Add-on", "Remove Supported Add-ons", "Change GW2 install Path: [" + installDir + "]"},
		cursor:   0,
		selected: make(map[int]struct{}),				
	}

}

func (m model) Init() tea.Cmd {
	return tea.SetWindowTitle("arcDPS Installer")
}


func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
        switch msg := msg.(type) {
		case errMsg: 
			m.err = msg.err				
		case statusMsg:
				m.status = string(msg)
        case tea.KeyMsg:
                switch msg.String() {
                case "ctrl+c", "q":
						m.quitting = true
                        return m, tea.Quit
                case "up", "j":
						if m.err != nil {m.err = nil} //clear error msg
                        if m.cursor > 0 {
                                m.cursor--
                        }
						m.err = nil
                case "down", "k":
						if m.err != nil {m.err = nil} //clear error msg
                        if m.cursor < len(m.choices)-1 {
                                m.cursor++
                        }
				case "ctrl+z":
					//toggle infoview
					var cmd tea.Cmd
					if m.altscreen {						
						cmd = tea.ExitAltScreen
						
					} else {
						cmd = tea.EnterAltScreen
					}
					m.altscreen = !m.altscreen
					return m, cmd
                case "enter", " ":
                        // The user has pressed enter or space, so select the current choice
                        m.choice = m.choices[m.cursor]
						
						//toggle infoview
						var cmd1 tea.Cmd
						var cmd2 tea.Cmd
						if !m.altscreen  {
							cmd1 = tea.EnterAltScreen
							cmd2 = downloadFunc(int(m.cursor))							
							m.altscreen = !m.altscreen
							if m.cursor == 0 {
								var cmd3 tea.Cmd
								var cmd4 tea.Cmd
								var cmd5 tea.Cmd
								var cmd6 tea.Cmd								
								cmd3 = downloadFunc(int(1))
								cmd4 = downloadFunc(int(2))
								cmd5 = downloadFunc(int(3))
								cmd6 = checkAppUpdates(arcDPSInstaller_urlString)
								m.status = ""
								infoCache = ""								
								return m, tea.Sequence(cmd1, cmd2, tea.Batch( cmd3, cmd4, cmd5),  cmd6)
							}
							if m.cursor == 5 {
								
								//update gw2 installdir
								cmd1 = tea.EnterAltScreen
								//load config 
								config, err := loadConfig(configFilePath)
								if err == nil {	
								cmd2 = updateInstallPath(config, configFilePath)
								}
							}	
							m.status = ""
							m.err = nil						
							return m, tea.Batch(cmd1, cmd2)
						}                    
                       
                }//msg.String
		
					
        }//msg.(type)
        return m, nil
}

func (m model) View() string {
		
		if m.quitting {
			return "Bye!\n"
		}

		const (
			altscreenMode = " Status "
			inlineMode    = " Main Menu "
		)
		var uiString string = ""
		var mode string
		if m.altscreen {
			mode = altscreenMode
		} else {
			mode = inlineMode
		}
		//show errors above header 
		if m.err != nil && m.cursor != 5 {	uiString += string(m.err.Error()) + "\n"}
		
		
		
		
		if  mode == altscreenMode {
			// Handle downloading ui here
			uiString += keywordStyle.Render(mode) + "\n\n"
			if m.cursor == 5 {uiString += sucessStyle.Render("File Picker Open. Waiting...") + "\n"}
			if m.cursor == 0   {						
			uiString += infoCache
			}
			if m.cursor > 0 {
				//show errors
				if m.err != nil{uiString += string(m.err.Error()) + "\n"}
				uiString += m.status
			}					
			uiString +=  helpStyle.Render("\n ctrl-z: switch to Main Menu • q: exit") + "\n"	 					
			return uiString
		}
		if  mode == inlineMode{	
			if version == "" {
			//version is set at compile time so in could be empty during testing
			//versionInt will be set to 0.0.1
			version = "0.0.1"
			}
			//update UI with current installDir value
			if m.cursor == 5 {m.choices[5] = "Change GW2 install Path: [" + installDir + "]"}
			uiString += dlSeperatorStyle.Render("----------------------------------------------------------------------") + "\n"			
			uiString +=  dlSeperatorStyle.Render("| ") + sucessStyle.Render("Website => https://github.com/theextendedname/arcDPS-Installer") + dlSeperatorStyle.Render("     |")+ "\n"
			uiString +=  dlSeperatorStyle.Render("| ") +  sucessStyle.Render("install, update and remove arcDPS, Healing, and Boon-Table Add-ons") + dlSeperatorStyle.Render(" |") + "\n"
			uiString += dlSeperatorStyle.Render("----------------------------------------------------------------------") + "\n"
			uiString +=  keywordStyle.Render(mode) + "\n\n"
			for i, choice := range m.choices {//build main menu
					var cursor string
					if m.cursor == i {
							cursor = "> "
							uiString += fmt.Sprintf("%s "+ selestedStyle.Render("%s") + "\n", cursor, choice)
					} else {
							cursor = "  "
							uiString += fmt.Sprintf("%s %s\n", cursor, choice)
					}

			}
			 uiString += helpStyle.Render("\n Use ↑↓ Arrow or JK to select option. • Then Press Enter. • ctrl-z: switch to Status View • q: exit")  + "\n"
			 uiString += "\n" + sucessStyle.Render("arcDPS-Instaler Version ") + sucessStyle.Render(version) + "\n" + sucessStyle.Render("by TheExtendedName") + "\n"


			return uiString
		}
		
	return uiString
}

func main() {
	//returns the absolute path to the executable file itself
	exePath, err := os.Executable()
	if err != nil {
			log.Fatal("Error getting executable path:" , err)
	}
	//get the directory containing the executable
	exeDir := filepath.Dir(exePath)
	configFilePath = exeDir +"/arcDPS-Installer-Config.json" // config path
	//load config 
	config, err := loadConfig(configFilePath)
	if err == nil {	
		if config.GW2PathOverRide == "" {
			// no config found use what's in the registry
			//get install dir from registry
			installDirTmp , err := getInstallPath()			
			if err == nil &&  directoryExists(installDirTmp){		
			//set install dir from registry
			installDir = installDirTmp
			}
			//if there is an error the installDir has the default set by global var
		}else{
		//config found use it
		installDir = config.GW2PathOverRide
		}
	}//if there is an error the installDir has the default set by global var
		
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
	}
		
	
}