package config

import "github.com/adrg/xdg"

const (
	App              string = "xrc"
	Version          string = "v0.1.0"
	ReleaseLatestURL string = "https://github.com/drawmoon/xrc/releases/latest"
)

var (
	UserHomeDir  = xdg.Home
	AppConfigDir = xdg.ConfigHome + "/" + App
	AppCacheDir  = xdg.CacheHome + "/" + App
	AppDataDir   = xdg.DataHome + "/" + App
	AppStateDir  = xdg.StateHome + "/" + App
)

var (
	ConfigFile  string = AppConfigDir + "/config.toml"
	HotkeysFile string = AppConfigDir + "/hotkeys.toml"
	ThemeFolder string = AppConfigDir + "/theme"

	ToggleDotFile string = AppDataDir + "/toggleDotFile"

	FixHotkeys    bool   = false
	FixConfigFile bool   = false
	LastDir       string = ""
	PrintLastDir  bool   = false
)

const (
	TrashDirectory      string = "/Trash"
	TrashDirectoryFiles string = "/Trash/files"
	TrashDirectoryInfo  string = "/Trash/info"
)
