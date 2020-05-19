package cmd

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configSourceDirKey      string      = "source_dir"
	configSourceDirLocalKey string      = "source_dir_local"
	rootExecutionRelPath    string      = "."
	configFileName          string      = "lefthook"
	configLocalFileName     string      = "lefthook-local"
	configExtendsOption     string      = "extends"
	configExtension         string      = ".yml"
	configExtensionPattern  string      = ".*"
	defaultFilePermission   os.FileMode = 0755
	defaultDirPermission    os.FileMode = 0666
)

var (
	Verbose              bool
	NoColors             bool
	rootPath             string
	cfgFile              string
	originConfig         *viper.Viper
	configFileExtensions = []string{".yml", ".yaml"}

	au aurora.Aurora
)

var rootCmd = &cobra.Command{
	Use:   "lefthook",
	Short: "CLI tool to manage Git hooks",
	Long: `After installation go to your project directory
and execute the following command:
lefthook install`,
}

func Execute() {
	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			os.Exit(1)
		}
	}()

	err := rootCmd.Execute()
	check(err)
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVar(&NoColors, "no-colors", false, "disable colored output")

	initAurora()
	cobra.OnInitialize(initConfig)
	// re-init Aurora after config reading because `colors` can be specified in config
	cobra.OnInitialize(initAurora)
	log.SetOutput(os.Stdout)
}

func initAurora() {
	au = aurora.NewAurora(EnableColors())
}

func initConfig() {
	log.SetFlags(0)

	setRootPath(rootExecutionRelPath)

	// store original config before merge
	originConfig = viper.New()
	originConfig.SetConfigName(configFileName)
	originConfig.AddConfigPath(rootExecutionRelPath)
	originConfig.ReadInConfig()

	viper.SetConfigName(configFileName)
	viper.AddConfigPath(rootExecutionRelPath)
	viper.SetDefault(configSourceDirKey, ".lefthook")
	viper.SetDefault(configSourceDirLocalKey, ".lefthook-local")
	viper.ReadInConfig()

	viper.SetConfigName(configLocalFileName)
	viper.MergeInConfig()

	if isConfigExtends() {
		for _, path := range getExtendsPath() {
			filename := filepath.Base(path)
			extension := filepath.Ext(path)
			name := filename[0 : len(filename)-len(extension)]
			viper.SetConfigName(name)
			viper.AddConfigPath(filepath.Dir(path))
			err := viper.MergeInConfig()
			check(err)
		}
	}

	viper.AutomaticEnv()
}

func getRootPath() string {
	return rootPath
}

func setRootPath(path string) {
	// get absolute path to .git dir (project root)
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")

	outputBytes, err := cmd.CombinedOutput()
	check(err)
	rootPath = strings.TrimSpace(string(outputBytes))
}

func getSourceDir() string {
	return filepath.Join(getRootPath(), viper.GetString(configSourceDirKey))
}

func getLocalSourceDir() string {
	return filepath.Join(getRootPath(), viper.GetString(configSourceDirLocalKey))
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func isConfigExtends() bool {
	return len(viper.GetStringSlice(configExtendsOption)) > 0
}

func getExtendsPath() []string {
	return viper.GetStringSlice(configExtendsOption)
}

// EnableColors shows is colors supported for current output or not.
// If `colors` explicitly specified in config, will return this value.
// Otherwise enabled for TTY and disabled for non-terminal output.
func EnableColors() bool {
	if NoColors {
		return false
	}
	if !viper.IsSet(colorsConfigKey) {
		return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	}
	return viper.GetBool(colorsConfigKey)
}

// VerbosePrint print text if Verbose flag persist
func VerbosePrint(v ...interface{}) {
	if Verbose {
		log.Println(v...)
	}
}
