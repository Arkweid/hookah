// On windows we can`t use pty, so we cant sync output in parallel mode

package cmd

import (
	"errors"
	"fmt"

	// "io" // win specific
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/Arkweid/lefthook/context"

	arrop "github.com/adam-hanna/arrayOperations"
	"github.com/gobwas/glob"

	// "github.com/kr/pty" //win specific
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/alessio/shellescape.v1"
)

var (
	hooksGroup     string
	executableName string
	sourcePath     string
	okList         []string
	failList       []string
	mutex          sync.Mutex
	envExcludeTags []string // store for LEFTHOOK_EXCLUDE=tag,tag
	isPipeBroken   bool
)

const (
	rootConfigKey        string      = "root"
	runnerConfigKey      string      = "runner"
	runConfigKey         string      = "run" // alias for runner
	runnerArgsConfigKey  string      = "runner_args"
	scriptsConfigKey     string      = "scripts"
	commandsConfigKey    string      = "commands"
	includeConfigKey     string      = "include"
	excludeConfigKey     string      = "exclude"
	globConfigKey        string      = "glob"
	skipConfigKey        string      = "skip"
	skipEmptyConfigKey   string      = "skip_empty"
	filesConfigKey       string      = "files"
	colorsConfigKey      string      = "colors"
	parallelConfigKey    string      = "parallel"
	subFiles             string      = "{files}"
	subAllFiles          string      = "{all_files}"
	subStagedFiles       string      = "{staged_files}"
	pushFiles            string      = "{push_files}"
	runnerWrapPattern    string      = "{cmd}"
	tagsConfigKey        string      = "tags"
	pipedConfigKey       string      = "piped"
	excludeTagsConfigKey string      = "exclude_tags"
	execMode             os.FileMode = 0751
	stageFixedConfigKey  string      = "stage_fixed"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute group of hooks",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		err := RunCmdExecutor(args, appFs)
		if err != nil {
			os.Exit(1)
		}
		os.Exit(0)
	},
}

func init() {
	runCmd.SetUsageTemplate(`Usage:
    lefthook run [hooksGroup]
Example:
    lefthook run pre-commit
`)
	rootCmd.AddCommand(runCmd)
}

// RunCmdExecutor run executables in hook groups
func RunCmdExecutor(args []string, fs afero.Fs) error {
	if os.Getenv("LEFTHOOK") == "0" {
		return nil
	}
	if tags := os.Getenv("LEFTHOOK_EXCLUDE"); tags != "" {
		envExcludeTags = append(envExcludeTags, strings.Split(tags, ",")[:]...)
	}

	hooksGroup := args[0]
	if !viper.IsSet(hooksGroup) && hooksGroup == "prepare-commit-msg" {
		return nil
	}
	gitArgs := args[1:]
	var wg sync.WaitGroup

	startTime := time.Now()
	log.Println(au.Cyan("RUNNING HOOKS GROUP:"), au.Bold(hooksGroup))

	if isPipedAndParallel(hooksGroup) {
		log.Println(au.Brown("Config error! Conflicted options 'piped' and 'parallel'. Remove one of this option from hook group."))
		return errors.New("Piped and Parallel options in conflict")
	}

	sourcePath := filepath.Join(getSourceDir(), hooksGroup)
	executables, err := afero.ReadDir(fs, sourcePath)
	if err == nil && len(executables) > 0 {
		for _, executable := range executables {
			wg.Add(1)
			if getParallel(hooksGroup) {
				go executeScript(hooksGroup, sourcePath, executable, &wg, gitArgs)
			} else {
				executeScript(hooksGroup, sourcePath, executable, &wg, gitArgs)
			}
		}
	}

	sourcePath = filepath.Join(getLocalSourceDir(), hooksGroup)
	executables, err = afero.ReadDir(fs, sourcePath)
	if err == nil && len(executables) > 0 {
		for _, executable := range executables {
			wg.Add(1)
			if getParallel(hooksGroup) {
				go executeScript(hooksGroup, sourcePath, executable, &wg, gitArgs)
			} else {
				executeScript(hooksGroup, sourcePath, executable, &wg, gitArgs)
			}
		}
	}

	commands := getCommands(hooksGroup)
	if len(commands) != 0 {

		for _, commandName := range commands {
			wg.Add(1)
			if getParallel(hooksGroup) {
				go executeCommand(hooksGroup, commandName, &wg)
			} else {
				executeCommand(hooksGroup, commandName, &wg)
			}
		}
	}

	wg.Wait()

	printSummary(time.Now().Sub(startTime))

	if len(failList) == 0 {
		return nil
	}
	return errors.New("Have failed script")
}

func executeCommand(hooksGroup, commandName string, wg *sync.WaitGroup) {
	defer wg.Done()

	if getPiped(hooksGroup) && isPipeBroken {
		mutex.Lock()
		log.Println("\n", au.Bold(commandName), au.Brown("(SKIP BY BROKEN PIPE)"))
		mutex.Unlock()
		return
	}

	files := []string{}
	runner := getRunner(hooksGroup, commandsConfigKey, commandName)

	if strings.Contains(runner, subStagedFiles) {
		files, _ = context.StagedFiles()
	} else if strings.Contains(runner, subFiles) || getCommandFiles(hooksGroup, commandName) != "" {
		files, _ = context.ExecGitCommand(getCommandFiles(hooksGroup, commandName))
	} else if strings.Contains(runner, pushFiles) {
		files, _ = context.PushFiles()
	} else {
		files, _ = context.AllFiles()
	}

	VerbosePrint("Files before filters: \n", files)

	files = FilterGlob(files, getCommandGlobRegexp(hooksGroup, commandName))
	files = FilterInclude(files, getCommandIncludeRegexp(hooksGroup, commandName)) // NOTE: confusing option, suppose delete it
	files = FilterExclude(files, getCommandExcludeRegexp(hooksGroup, commandName))

	cmdRoot := getRoot(hooksGroup, commandName)
	files = FilterRelative(files, cmdRoot)

	VerbosePrint("Files after filters: \n", files)

	files_esc := []string{}
	for _, fileName := range files {
		if len(fileName) > 0 {
			files_esc = append(files_esc, shellescape.Quote(fileName))
		}
	}
	files = files_esc
	VerbosePrint("Files after escaping: \n", files)

	runner = strings.Replace(runner, pushFiles, strings.Join(files, " "), -1)
	runner = strings.Replace(runner, subStagedFiles, strings.Join(files, " "), -1)
	runner = strings.Replace(runner, subAllFiles, strings.Join(files, " "), -1)
	runner = strings.Replace(runner, subFiles, strings.Join(files, " "), -1)

	if isSkipCommmand(hooksGroup, commandName) {
		mutex.Lock()
		log.Println("\n", au.Bold(commandName), au.Brown("(SKIP BY SETTINGS)"))
		mutex.Unlock()
		return
	}
	if result, _ := arrop.Intersect(getExcludeTags(hooksGroup), getTags(hooksGroup, commandsConfigKey, commandName)); len(result.Interface().([]string)) > 0 {
		mutex.Lock()
		log.Println("\n", au.Bold(commandName), au.Brown("(SKIP BY TAGS)"))
		mutex.Unlock()
		return
	}
	if len(files) < 1 && isSkipEmptyCommmand(hooksGroup, commandName) {
		mutex.Lock()
		log.Println("\n", au.Bold(commandName), au.Brown("(SKIP. NO FILES FOR INSPECTING)"))
		mutex.Unlock()
		return
	}

	runnerArg := strings.Split(runner, " ")
	command := exec.Command(runnerArg[0], runnerArg[1:]...)
	if cmdRoot != "" {
		fullPath, _ := filepath.Abs(cmdRoot)
		command.Dir = fullPath
	}
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout // win specific
	command.Stderr = os.Stderr // win specific

	err := command.Start() // ptyOut, err := pty.Start(command) // win specific
	mutex.Lock()
	defer mutex.Unlock()

	log.Println(au.Cyan("\n  EXECUTE >"), au.Bold(commandName))
	if err != nil {
		processError(commandName, err)
		return
	}
	fileNames, err := command.Output()
	if err != nil {
		processError(commandName, err)
		return
	}
	// io.Copy(os.Stdout, ptyOut) // win specific
	if command.Wait() == nil {
		okList = append(okList, commandName)
		context.ExecGitCommand(fmt.Sprintf("git add %v", strings.Replace(string(fileNames), "\r\n", " ", -1)))

	} else {
		failList = append(failList, commandName)
		setPipeBroken()

	}
}

func executeScript(hooksGroup, source string, executable os.FileInfo, wg *sync.WaitGroup, gitArgs []string) {
	defer wg.Done()
	executableName := executable.Name()

	if getPiped(hooksGroup) && isPipeBroken {
		mutex.Lock()
		log.Println("\n", au.Bold(executableName), au.Brown("(SKIP BY BROKEN PIPE)"))
		mutex.Unlock()
		return
	}

	pathToExecutable := filepath.Join(source, executableName)

	if err := isExecutable(executable); err != nil {
		makeExecutable(pathToExecutable)
	}

	pathToExecutable, _ = filepath.Rel(getRootPath(), pathToExecutable)

	command := exec.Command(pathToExecutable, gitArgs[:]...)

	if haveRunner(hooksGroup, scriptsConfigKey, executableName) {
		runnerArg := strings.Split(getRunner(hooksGroup, scriptsConfigKey, executableName), " ")
		runnerArg = append(runnerArg, pathToExecutable)
		runnerArg = append(runnerArg, gitArgs[:]...)

		command = exec.Command(runnerArg[0], runnerArg[1:]...)
	}
	command.Stdin = os.Stdin
	command.Stdout = os.Stdout // win specific
	command.Stderr = os.Stderr // win specific

	if !isScriptExist(hooksGroup, executableName) {
		mutex.Lock()
		log.Println("\n", au.Bold(executableName), au.Brown("(SKIP BY NOT EXIST IN CONFIG)"))
		mutex.Unlock()
		return
	}
	if isSkipScript(hooksGroup, executableName) {
		mutex.Lock()
		log.Println("\n", au.Bold(executableName), au.Brown("(SKIP BY SETTINGS)"))
		mutex.Unlock()
		return
	}
	if result, _ := arrop.Intersect(getExcludeTags(hooksGroup), getTags(hooksGroup, scriptsConfigKey, executableName)); len(result.Interface().([]string)) > 0 {
		mutex.Lock()
		log.Println("\n", au.Bold(executableName), au.Brown("(SKIP BY TAGS)"))
		mutex.Unlock()
		return
	}

	err := command.Start() // ptyOut, err := pty.Start(command) // win specific
	mutex.Lock()
	defer mutex.Unlock()

	log.Println(au.Cyan("\n  EXECUTE >"), au.Bold(executableName))
	if os.IsPermission(err) {
		log.Println(au.Brown("(SKIP NOT EXECUTABLE FILE)"))
		return
	}
	if err != nil {
		failList = append(failList, executableName)
		log.Println(err)
		log.Println(au.Brown("TIP: Command start failed. Checkout `runner:` option for this script"))
		setPipeBroken()
		return
	}

	// io.Copy(os.Stdout, ptyOut) // win specific
	if command.Wait() == nil {
		okList = append(okList, executableName)
	} else {
		failList = append(failList, executableName)
		setPipeBroken()
	}
}

func haveRunner(hooksGroup, source, executableName string) (out bool) {
	if runner := getRunner(hooksGroup, source, executableName); runner != "" {
		out = true
	}
	return
}

func getRunner(hooksGroup, source, executableName string) string {
	key := strings.Join([]string{hooksGroup, source, executableName, runnerConfigKey}, ".")
	runner := viper.GetString(key)

	aliasKey := strings.Join([]string{hooksGroup, source, executableName, runConfigKey}, ".")
	aliasRunner := viper.GetString(aliasKey)
	if runner == "" && aliasRunner != "" {
		runner = aliasRunner
	}

	// If runner have {cmd} substring, replace it from runner in lefthook.yaml
	if res := strings.Contains(runner, runnerWrapPattern); res {
		originRunner := originConfig.GetString(key)
		runner = strings.Replace(runner, runnerWrapPattern, originRunner, -1)
	}

	return runner
}

func printSummary(execTime time.Duration) {
	if len(okList) == 0 && len(failList) == 0 {
		log.Println(au.Cyan("\nSUMMARY:"), au.Brown("(SKIP EMPTY)"))
	} else {
		log.Println(au.Cyan(fmt.Sprintf("\nSUMMARY: (done in %.2f seconds)", execTime.Seconds())))
	}

	for _, fileName := range okList {
		log.Printf("✔️  %s\n", au.Green(fileName))
	}

	for _, fileName := range failList {
		log.Printf("🥊  %s", au.Red(fileName))
	}
}

func isScriptExist(hooksGroup, executableName string) bool {
	key := strings.Join([]string{hooksGroup, scriptsConfigKey, executableName}, ".")
	return viper.IsSet(key)
}

func isStageFixedFiles(hooksGroup, executableName string) bool {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, stageFixedConfigKey}, ".")
	return viper.GetBool(key)
}

func isSkipScript(hooksGroup, executableName string) bool {
	key := strings.Join([]string{hooksGroup, scriptsConfigKey, executableName, skipConfigKey}, ".")
	return viper.GetBool(key)
}

func isSkipCommmand(hooksGroup, executableName string) bool {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, skipConfigKey}, ".")
	return viper.GetBool(key)
}

// NOTE: confusing option, suppose it unnesecary and should be deleted.
func isSkipEmptyCommmand(hooksGroup, executableName string) bool {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, skipEmptyConfigKey}, ".")
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}

	key = strings.Join([]string{hooksGroup, skipEmptyConfigKey}, ".")
	if viper.IsSet(key) {
		return viper.GetBool(key)
	}

	return true
}

func getCommands(hooksGroup string) []string {
	key := strings.Join([]string{hooksGroup, commandsConfigKey}, ".")
	commands := viper.GetStringMap(key)

	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func getRoot(hooksGroup string, executableName string) string {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, rootConfigKey}, ".")
	return viper.GetString(key)
}

func getCommandIncludeRegexp(hooksGroup, executableName string) string {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, includeConfigKey}, ".")
	return viper.GetString(key)
}

func getCommandExcludeRegexp(hooksGroup, executableName string) string {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, excludeConfigKey}, ".")
	return viper.GetString(key)
}

func getCommandGlobRegexp(hooksGroup, executableName string) string {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, globConfigKey}, ".")
	if viper.GetString(key) != "" {
		return viper.GetString(key)
	}

	key = strings.Join([]string{hooksGroup, globConfigKey}, ".")
	return viper.GetString(key)
}

func getCommandFiles(hooksGroup, executableName string) string {
	key := strings.Join([]string{hooksGroup, commandsConfigKey, executableName, filesConfigKey}, ".")
	if viper.GetString(key) != "" {
		return viper.GetString(key)
	}

	key = strings.Join([]string{hooksGroup, filesConfigKey}, ".")
	if viper.GetString(key) != "" {
		return viper.GetString(key)
	}

	return ""
}

func getTags(hooksGroup, source, executableName string) []string {
	key := strings.Join([]string{hooksGroup, source, executableName, tagsConfigKey}, ".")
	return strings.Split(viper.GetString(key), " ")
}

func getExcludeTags(hooksGroup string) []string {
	key := strings.Join([]string{hooksGroup, excludeTagsConfigKey}, ".")
	if len(viper.GetStringSlice(key)) > 0 {
		return append(viper.GetStringSlice(key), envExcludeTags[:]...)
	}

	if len(viper.GetStringSlice(excludeTagsConfigKey)) > 0 {
		return append(viper.GetStringSlice(excludeTagsConfigKey), envExcludeTags[:]...)
	}

	if len(envExcludeTags) > 0 {
		return envExcludeTags
	}

	return []string{}
}

func getParallel(hooksGroup string) bool {
	key := strings.Join([]string{hooksGroup, parallelConfigKey}, ".")
	return viper.GetBool(key)
}

func getPiped(hooksGroup string) bool {
	key := strings.Join([]string{hooksGroup, pipedConfigKey}, ".")
	return viper.GetBool(key)
}

func isPipedAndParallel(hooksGroup string) bool {
	return getParallel(hooksGroup) && getPiped(hooksGroup)
}

func setPipeBroken() {
	isPipeBroken = true
}

func FilterGlob(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	g := glob.MustCompile(matcher)

	vsf := make([]string, 0)
	for _, v := range vs {
		if res := g.Match(v); res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterRelative(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	vsf := make([]string, 0)
	for _, v := range vs {
		if strings.HasPrefix(v, matcher) {
			vsf = append(vsf, strings.Replace(v, matcher, "./", 1))
		}
	}
	return vsf
}

func FilterInclude(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	vsf := make([]string, 0)
	for _, v := range vs {
		if res, _ := regexp.MatchString(matcher, v); res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func FilterExclude(vs []string, matcher string) []string {
	if matcher == "" {
		return vs
	}

	vsf := make([]string, 0)
	for _, v := range vs {
		if res, _ := regexp.MatchString(matcher, v); !res {
			vsf = append(vsf, v)
		}
	}
	return vsf
}

func isExecutable(executable os.FileInfo) error {
	mode := executable.Mode()

	if !mode.IsRegular() {
		return errors.New("ErrPermission")
	}
	if (mode & 0111) == 0 {
		return errors.New("ErrPermission")
	}
	return nil
}

func makeExecutable(path string) {
	if err := os.Chmod(path, execMode); err != nil {
		log.Fatal(err)
	}
}

func processError(commandName string, err error) {
	failList = append(failList, commandName)
	setPipeBroken()
	log.Println(err)
}
