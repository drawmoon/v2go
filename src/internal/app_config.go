package internal

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/barasher/go-exiftool"
	"github.com/charmbracelet/lipgloss"
	"github.com/pelletier/go-toml/v2"
	variable "github.com/yorukot/superfile/src/config"
	"github.com/yorukot/superfile/src/config/icon"
)

var (
	wheelRunTime = 5
)

var (
	footerBorderColor       lipgloss.Color
	footerBorderActiveColor lipgloss.Color

	fullScreenBGColor lipgloss.Color
	filePanelBGColor  lipgloss.Color
	sidebarBGColor    lipgloss.Color
	footerBGColor     lipgloss.Color
	modalBGColor      lipgloss.Color

	fullScreenFGColor lipgloss.Color
	filePanelFGColor  lipgloss.Color
	sidebarFGColor    lipgloss.Color
	footerFGColor     lipgloss.Color
	modalFGColor      lipgloss.Color

	cursorColor  lipgloss.Color
	correctColor lipgloss.Color
	errorColor   lipgloss.Color
	hintColor    lipgloss.Color
	cancelColor  lipgloss.Color

	filePanelTopDirectoryIconColor lipgloss.Color
	filePanelTopPathColor          lipgloss.Color
	filePanelItemSelectedFGColor   lipgloss.Color
	filePanelItemSelectedBGColor   lipgloss.Color

	sidebarTitleColor          lipgloss.Color
	sidebarItemSelectedFGColor lipgloss.Color
	sidebarItemSelectedBGColor lipgloss.Color
	sidebarDividerColor        lipgloss.Color

	modalCancelFGColor  lipgloss.Color
	modalCancelBGColor  lipgloss.Color
	modalConfirmFGColor lipgloss.Color
	modalConfirmBGColor lipgloss.Color

	helpMenuHotkeyColor lipgloss.Color
	helpMenuTitleColor  lipgloss.Color
)

// Variables for holding default configurations of each settings
var (
	HotkeysTomlString  string
	ConfigTomlString   string
	DefaultThemeString string
)

// Configuration settings
type AppConfig struct {
	Theme string `toml:"theme" comment:"More details are at https://superfile.netlify.app/configure/superfile-config/\nchange your theme"`

	Editor                 string `toml:"editor" comment:"\nThe editor files/directories will be opened with. (leave blank to use the EDITOR environment variable)."`
	AutoCheckUpdate        bool   `toml:"auto_check_update" comment:"\nAuto check for update"`
	CdOnQuit               bool   `toml:"cd_on_quit" comment:"\nCd on quit (For more details, please check out https://superfile.netlify.app/configure/superfile-config/#cd_on_quit)"`
	DefaultOpenFilePreview bool   `toml:"default_open_file_preview" comment:"\nWhether to open file preview automatically every time superfile is opened."`
	DefaultDirectory       string `toml:"default_directory" comment:"\nThe path of the first file panel when superfile is opened."`
	FileSizeUseSI          bool   `toml:"file_size_use_si" comment:"\nDisplay file sizes using powers of 1000 (kB, MB, GB) instead of powers of 1024 (KiB, MiB, GiB)."`
	DefaultSortType        int    `toml:"default_sort_type" comment:"\nDefault sort type (0: Name, 1: Size, 2: Date Modified)."`
	SortOrderReversed      bool   `toml:"sort_order_reversed" comment:"\nDefault sort order (false: Ascending, true: Descending)."`
	CaseSensitiveSort      bool   `toml:"case_sensitive_sort" comment:"\nCase sensitive sort by name (captal \"B\" comes before \"a\" if true)."`

	Nerdfont              bool `toml:"nerdfont" comment:"\n================   Style =================\n\n If you don't have or don't want Nerdfont installed you can turn this off"`
	TransparentBackground bool `toml:"transparent_background" comment:"\nSet transparent background or not (this only work when your terminal background is transparent)"`
	FilePreviewWidth      int  `toml:"file_preview_width" comment:"\nFile preview width allow '0' (this mean same as file panel),'x' x must be less than 10 and greater than 1 (This means that the width of the file preview will be one xth of the total width.)"`
	SidebarWidth          int  `toml:"sidebar_width" comment:"\nThe length of the sidebar. If you don't want to display the sidebar, you can input 0 directly. If you want to display the value, please place it in the range of 3-20."`

	BorderTop         string `toml:"border_top" comment:"\nBorder style"`
	BorderBottom      string `toml:"border_bottom"`
	BorderLeft        string `toml:"border_left"`
	BorderRight       string `toml:"border_right"`
	BorderTopLeft     string `toml:"border_top_left"`
	BorderTopRight    string `toml:"border_top_right"`
	BorderBottomLeft  string `toml:"border_bottom_left"`
	BorderBottomRight string `toml:"border_bottom_right"`
	BorderMiddleLeft  string `toml:"border_middle_left"`
	BorderMiddleRight string `toml:"border_middle_right"`

	Metadata          bool `toml:"metadata" comment:"\n==========PLUGINS========== #\n\nShow more detailed metadata, please install exiftool before enabling this plugin!"`
	EnableMD5Checksum bool `toml:"enable_md5_checksum" comment:"Enable MD5 checksum generation for files"`
}

// Theme configuration
type ThemeConfig struct {
	// Code syntax highlight theme
	CodeSyntaxHighlightTheme string `toml:"code_syntax_highlight"`

	// Border
	FilePanelBorder string `toml:"file_panel_border"`
	SidebarBorder   string `toml:"sidebar_border"`
	FooterBorder    string `toml:"footer_border"`

	// Border Active
	FilePanelBorderActive string `toml:"file_panel_border_active"`
	SidebarBorderActive   string `toml:"sidebar_border_active"`
	FooterBorderActive    string `toml:"footer_border_active"`
	ModalBorderActive     string `toml:"modal_border_active"`

	// Background (bg)
	FullScreenBG string `toml:"full_screen_bg"`
	FilePanelBG  string `toml:"file_panel_bg"`
	SidebarBG    string `toml:"sidebar_bg"`
	FooterBG     string `toml:"footer_bg"`
	ModalBG      string `toml:"modal_bg"`

	// Foreground (fg)
	FullScreenFG string `toml:"full_screen_fg"`
	FilePanelFG  string `toml:"file_panel_fg"`
	SidebarFG    string `toml:"sidebar_fg"`
	FooterFG     string `toml:"footer_fg"`
	ModalFG      string `toml:"modal_fg"`

	// Special Color
	Cursor        string   `toml:"cursor"`
	Correct       string   `toml:"correct"`
	Error         string   `toml:"error"`
	Hint          string   `toml:"hint"`
	Cancel        string   `toml:"cancel"`
	GradientColor []string `toml:"gradient_color"`

	// File Panel Special Items
	FilePanelTopDirectoryIcon string `toml:"file_panel_top_directory_icon"`
	FilePanelTopPath          string `toml:"file_panel_top_path"`
	FilePanelItemSelectedFG   string `toml:"file_panel_item_selected_fg"`
	FilePanelItemSelectedBG   string `toml:"file_panel_item_selected_bg"`

	// Sidebar Special Items
	SidebarTitle          string `toml:"sidebar_title"`
	SidebarItemSelectedFG string `toml:"sidebar_item_selected_fg"`
	SidebarItemSelectedBG string `toml:"sidebar_item_selected_bg"`
	SidebarDivider        string `toml:"sidebar_divider"`

	// Modal Special Items
	ModalCancelFG  string `toml:"modal_cancel_fg"`
	ModalCancelBG  string `toml:"modal_cancel_bg"`
	ModalConfirmFG string `toml:"modal_confirm_fg"`
	ModalConfirmBG string `toml:"modal_confirm_bg"`

	HelpMenuHotkey string `toml:"help_menu_hotkey"`
	HelpMenuTitle  string `toml:"help_menu_title"`
}

// Hotkeys configuration
type HotkeysConfig struct {
	Confirm []string `toml:"confirm" comment:"=================================================================================================\nGlobal hotkeys (cannot conflict with other hotkeys)"`
	Quit    []string `toml:"quit"`
	// movement
	ListUp   []string `toml:"list_up" comment:"movement"`
	ListDown []string `toml:"list_down"`
	PageUp   []string `toml:"page_up"`
	PageDown []string `toml:"page_down"`

	CloseFilePanel         []string `toml:"close_file_panel" comment:"file panel control"`
	CreateNewFilePanel     []string `toml:"create_new_file_panel"`
	NextFilePanel          []string `toml:"next_file_panel"`
	PreviousFilePanel      []string `toml:"previous_file_panel"`
	ToggleFilePreviewPanel []string `toml:"toggle_file_preview_panel"`
	OpenSortOptionsMenu    []string `toml:"open_sort_options_menu"`
	ToggleReverseSort      []string `toml:"toggle_reverse_sort"`

	FocusOnProcessBar []string `toml:"focus_on_process_bar" comment:"change focus"`
	FocusOnSidebar    []string `toml:"focus_on_sidebar"`
	FocusOnMetaData   []string `toml:"focus_on_metadata"`

	FilePanelItemCreate []string `toml:"file_panel_item_create" comment:"create file/directory and rename "`
	FilePanelItemRename []string `toml:"file_panel_item_rename"`

	CopyItems   []string `toml:"copy_items" comment:"file operate"`
	PasteItems  []string `toml:"paste_items"`
	CutItems    []string `toml:"cut_items"`
	DeleteItems []string `toml:"delete_items"`

	ExtractFile  []string `toml:"extract_file" comment:"compress and extract"`
	CompressFile []string `toml:"compress_file"`

	OpenFileWithEditor             []string `toml:"open_file_with_editor" comment:"editor"`
	OpenCurrentDirectoryWithEditor []string `toml:"open_current_directory_with_editor"`

	ToggleDotFile   []string `toml:"toggle_dot_file"`
	ChangePanelMode []string `toml:"change_panel_mode"`
	OpenHelpMenu    []string `toml:"open_help_menu"`
	OpenCommandLine []string `toml:"open_command_line"`

	CopyPath []string `toml:"copy_path"`
	CopyPWD  []string `toml:"copy_present_working_directory"`

	ConfirmTyping []string `toml:"confirm_typing" comment:"=================================================================================================\nTyping hotkeys (can conflict with all hotkeys)"`
	CancelTyping  []string `toml:"cancel_typing"`

	ParentDirectory []string `toml:"parent_directory" comment:"=================================================================================================\nNormal mode hotkeys (can conflict with other modes, cannot conflict with global hotkeys)"`
	SearchBar       []string `toml:"search_bar"`

	FilePanelSelectModeItemsSelectDown []string `toml:"file_panel_select_mode_items_select_down" comment:"=================================================================================================\nSelect mode hotkeys (can conflict with other modes, cananot conflict with global hotkeys)"`
	FilePanelSelectModeItemsSelectUp   []string `toml:"file_panel_select_mode_items_select_up"`
	FilePanelSelectAllItem             []string `toml:"file_panel_select_all_items"`
}

// Create proper directories for storing configuration and write default
// configurations to Config and Hotkeys toml
func InitConfigFile(content embed.FS) {
	// Load all default configurations from superfile_config folder
	loadDefaultConfig(content)

	// Create directories
	if err := createDirectories(
		variable.AppConfigDir,
		variable.AppDataDir,
		variable.AppStateDir,
		variable.ThemeFolder,
	); err != nil {
		log.Fatalln("Error creating directories:", err)
	}

	// Create files
	if err := createFiles(variable.ToggleDotFile); err != nil {
		log.Fatalln("Error creating files:", err)
	}

	// Write config file
	if err := writeConfigFile(variable.ConfigFile, ConfigTomlString); err != nil {
		log.Fatalln("Error writing config file:", err)
	}

	if err := writeConfigFile(variable.HotkeysFile, HotkeysTomlString); err != nil {
		log.Fatalln("Error writing config file:", err)
	}
}

// Load all default configurations from superfile_config folder into global
// configurations variables
func loadDefaultConfig(content embed.FS) {
	temp, err := content.ReadFile("src/superfile_config/hotkeys.toml")
	if err != nil {
		return
	}
	HotkeysTomlString = string(temp)

	temp, err = content.ReadFile("src/superfile_config/config.toml")
	if err != nil {
		return
	}
	ConfigTomlString = string(temp)

	temp, err = content.ReadFile("src/superfile_config/theme/amuse.toml")
	if err != nil {
		return
	}
	DefaultThemeString = string(temp)

	_, err = os.Stat(variable.ThemeFolder)
	if os.IsNotExist(err) {
		err := os.MkdirAll(variable.ThemeFolder, 0755)
		if err != nil {
			outPutLog("error create theme direcroty", err)
			return
		}
	}

	themeFiles, err := content.ReadDir("src/superfile_config/theme")
	if err != nil {
		outPutLog("error read theme directory from embed", err)
		return
	}
	for _, file := range themeFiles {
		if file.IsDir() {
			continue
		}
		src, err := content.ReadFile(filepath.Join("src/superfile_config/theme", file.Name()))
		if err != nil {
			outPutLog("error read theme file from embed", err)
			return
		}

		file, err := os.Create(filepath.Join(variable.ThemeFolder, file.Name()))
		if err != nil {
			outPutLog("error create theme file from embed", err)
			return
		}
		file.Write(src)
		defer file.Close()
	}
}

// Helper functions
// Create all dirs that does not already exists
func createDirectories(dirs ...string) error {
	for _, dir := range dirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			// Directory doesn't exist, create it
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
		} else if err != nil {
			// Some other error occurred while checking if the directory exists
			return fmt.Errorf("failed to check directory status %s: %w", dir, err)
		}
		// else: directory already exists
	}
	return nil
}

// Create all files if they do not exists yet
func createFiles(files ...string) error {
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			if err := os.WriteFile(file, nil, 0644); err != nil {
				return fmt.Errorf("failed to create file %s: %w", file, err)
			}
		}
	}
	return nil
}

// Write data to the path file if it exists
func writeConfigFile(path, data string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, []byte(data), 0644); err != nil {
			return fmt.Errorf("failed to write config file %s: %w", path, err)
		}
	}
	return nil
}

// initialConfig load and handle all configuration files (spf config,hotkeys
// themes) setted up. Returns absolute path of dir pointing to the file Panel
func initialConfig(dir string) (toggleDotFileBool bool, firstFilePanelDir string) {
	var err error

	loadConfigFile()
	loadHotkeysFile()
	loadThemeFile()
	icon.InitIcon(Config.Nerdfont)

	toggleDotFileData, err := os.ReadFile(variable.ToggleDotFile)
	if err != nil {
		outPutLog("Error while reading toggleDotFile data error:", err)
	}
	if string(toggleDotFileData) == "true" {
		toggleDotFileBool = true
	} else if string(toggleDotFileData) == "false" {
		toggleDotFileBool = false
	} else {
		toggleDotFileBool = false
	}

	LoadThemeConfig()

	if Config.Metadata {
		et, err = exiftool.NewExiftool()
		if err != nil {
			outPutLog("Initial model function init exiftool error", err)
		}
	}

	if dir != "" {
		firstFilePanelDir, err = filepath.Abs(dir)
	} else {
		Config.DefaultDirectory = strings.Replace(Config.DefaultDirectory, "~", variable.UserHomeDir, -1)
		firstFilePanelDir, err = filepath.Abs(Config.DefaultDirectory)
	}

	if err != nil {
		firstFilePanelDir = variable.UserHomeDir
	}

	return toggleDotFileBool, firstFilePanelDir
}

// Load configurations from the configuration file. Compares the content
// with the default values and modify the config file to include default configs
// if the FixConfigFile flag is on
func loadConfigFile() {

	//Initialize default configs
	_ = toml.Unmarshal([]byte(ConfigTomlString), &Config)
	//Initialize empty configs
	tempForCheckMissingConfig := AppConfig{}

	data, err := os.ReadFile(variable.ConfigFile)
	if err != nil {
		log.Fatalf("Config file doesn't exist: %v", err)
	}

	// Insert data present in the config file inside temp variable
	_ = toml.Unmarshal(data, &tempForCheckMissingConfig)
	// Replace default values for values specifieds in config file
	err = toml.Unmarshal(data, &Config)
	if err != nil && !variable.FixConfigFile {
		fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#F93939")).Render("Error") +
			lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFEE")).Render(" ┃ ") +
			"Error decoding configuration file\n")
		fmt.Println("To add missing fields to hotkeys directory automaticially run Superfile with the --fix-config-file flag `spf --fix-config-file`")
	}

	// If data is different and FixConfigFile option is on, then fullfill then
	// fullfill the config file with the default values
	if !reflect.DeepEqual(Config, tempForCheckMissingConfig) && variable.FixConfigFile {
		tomlData, err := toml.Marshal(Config)
		if err != nil {
			log.Fatalf("Error encoding config: %v", err)
		}

		err = os.WriteFile(variable.ConfigFile, tomlData, 0644)
		if err != nil {
			log.Fatalf("Error writing config file: %v", err)
		}
	}

	if (Config.FilePreviewWidth > 10 || Config.FilePreviewWidth < 2) && Config.FilePreviewWidth != 0 {
		fmt.Println(loadConfigError("file_preview_width"))
		os.Exit(0)
	}

	if Config.SidebarWidth != 0 && (Config.SidebarWidth < 3 || Config.SidebarWidth > 30) {
		fmt.Println(loadConfigError("sidebar_width"))
		os.Exit(0)
	}
}

// Load keybinds from the hotkeys file. Compares the content
// with the default values and modify the hotkeys  if the FixHotkeys flag is on.
// If is off check if all hotkeys are properly setted
func loadHotkeysFile() {
	// load default Hotkeys configs
	_ = toml.Unmarshal([]byte(HotkeysTomlString), &hotkeys)
	hotkeysFromConfig := HotkeysConfig{}
	data, err := os.ReadFile(variable.HotkeysFile)

	if err != nil {
		log.Fatalf("Config file doesn't exist: %v", err)
	}
	// Load data from hotkeys file
	_ = toml.Unmarshal(data, &hotkeysFromConfig)
	// Override default hotkeys with the ones from the file
	err = toml.Unmarshal(data, &hotkeys)
	if err != nil {
		log.Fatalf("Error decoding hotkeys file ( your config file may have misconfigured ): %v", err)
	}

	hasMissingHotkeysInConfig := !reflect.DeepEqual(hotkeys, hotkeysFromConfig)

	// If FixHotKeys is not on then check if every needed hotkey is properly setted
	if hasMissingHotkeysInConfig && !variable.FixHotkeys {
		hotKeysConfig := reflect.ValueOf(hotkeysFromConfig)
		for i := 0; i < hotKeysConfig.NumField(); i++ {
			field := hotKeysConfig.Type().Field(i)
			value := hotKeysConfig.Field(i)
			name := field.Name
			isMissing := value.Len() == 0

			if isMissing {
				fmt.Print(lipgloss.NewStyle().Foreground(lipgloss.Color("#F93939")).Render("Error") +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#00FFEE")).Render(" ┃ ") +
					fmt.Sprintf("Field \"%s\" is missing in hotkeys configuration\n", name))
			}
		}
		fmt.Println("To add missing fields to hotkeys directory automaticially run Superfile with the --fix-hotkeys flag `spf --fix-hotkeys`")
	}

	// Override hotkey files with default configs if the Fix flag is on
	if hasMissingHotkeysInConfig && variable.FixHotkeys {
		writeHotkeysFile(hotkeys)
	}

	val := reflect.ValueOf(hotkeys)

	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		value := val.Field(i)

		if value.Kind() != reflect.Slice || value.Type().Elem().Kind() != reflect.String {
			fmt.Println(lodaHotkeysError(field.Name))
			os.Exit(0)
		}

		hotkeysList := value.Interface().([]string)

		if len(hotkeysList) == 0 || hotkeysList[0] == "" {
			fmt.Println(lodaHotkeysError(field.Name))
			os.Exit(0)
		}
	}

}

// Write hotkeys inside the hotkeys toml file
func writeHotkeysFile(hotkeys HotkeysConfig) {
	tomlData, err := toml.Marshal(hotkeys)
	if err != nil {
		log.Fatalf("Error encoding hotkeys: %v", err)
	}

	err = os.WriteFile(variable.HotkeysFile, tomlData, 0644)
	if err != nil {
		log.Fatalf("Error writing hotkeys file: %v", err)
	}
}

// Load configurations from theme file into &theme and return default values
// if file theme folder is empty
func loadThemeFile() {
	data, err := os.ReadFile(variable.ThemeFolder + "/" + Config.Theme + ".toml")
	if err != nil {
		data = []byte(DefaultThemeString)
	}

	err = toml.Unmarshal(data, &theme)
	if err != nil {
		log.Fatalf("Error while decoding theme file( Your theme file may have errors ): %v", err)
	}
}
