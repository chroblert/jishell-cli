package cmd

import (
	"fmt"
	"github.com/chroblert/jishell"
	"github.com/chroblert/jlog"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

var addCmd = &jishell.Command{
	Name:      "add",
	Aliases:   nil,
	Help:      "add a command to a jishell Application",
	LongHelp:  "",
	HelpGroup: "",
	Usage:     "add [command name]",
	Flags: func(f *jishell.Flags) {
		f.String("t", "package", "", "target package name")
		f.String("p", "parent", "", "variable name of parent command for this comand")
	},
	Args: func(a *jishell.Args) {
		a.String("command", "sub command name.")
	},
	Run: func(c *jishell.Context) error {
		commandArg := c.Args.String("command")
		if len(commandArg) == 0 {
			return fmt.Errorf("请输入subCommand")
		}
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		commandName := validateCmdName(commandArg)
		cmdParent := c.Flags.String("parent")
		cmdPathList := filepath.SplitList(wd)
		cmdPathList = append(cmdPathList, "cmd")
		cmdPathList = append(cmdPathList, strings.Split(cmdParent, "/")...)
		cmdTplPrefix := ""
		cmdImportNamePrefix := ""
		modName := getModImportPath()
		cmdPkgName := ""
		if len(cmdParent) > 0 {
			cmdTplPrefix = strings.Join(strings.Split(cmdParent, "/"), "_")
			cmdImportNamePrefix = modName + "/cmd/" + cmdParent
			if strings.ContainsRune(cmdParent, '/') {
				cmdPkgName = cmdParent[strings.LastIndexByte(cmdParent, '/')+1:]
			} else {
				cmdPkgName = cmdParent
			}
		} else {
			cmdImportNamePrefix = modName + "/cmd"
			cmdPkgName = "cmd"
		}
		//jlog.Info("cmdParent:",cmdParent)
		//jlog.Info("modName:",modName)
		//jlog.Info("cmdTplPrefix:",cmdTplPrefix)
		//jlog.Info("cmdImportNamePrefix:",cmdImportNamePrefix)
		//jlog.Info(cmdImportNamePrefix)
		command := &Command{
			CmdName:             commandName,
			CmdParent:           cmdParent,
			CmdParentHandled:    strings.Join(strings.Split(cmdParent, "/"), "_"),
			CmdPath:             filepath.Join(cmdPathList...),
			CmdTplPrefix:        cmdTplPrefix,
			CmdImportNamePrefix: cmdImportNamePrefix,
			CmdPkgName:          cmdPkgName,
			Project: &Project{
				AbsolutePath: wd,
				Viper:        false,
			},
		}
		err = command.Create()
		if err != nil {
			return err
		}
		jlog.NInfof("%s created at %s\n", command.CmdName, command.AbsolutePath)
		return nil
	},
	Completer: nil,
	CMDPath:   "cmd",
}

func init() {
	var tmpCommands []*jishell.Command
	if viper.Get("jCommands") == nil {
		tmpCommands = make([]*jishell.Command, 0)
	} else {
		tmpCommands = viper.Get("jCommands").([]*jishell.Command)
	}
	tmpCommands = append(tmpCommands, addCmd)
	viper.Set("jCommands", tmpCommands)
}

// validateCmdName returns source without any dashes and underscore.
// If there will be dash or underscore, next letter will be uppered.
// It supports only ASCII (1-byte character) strings.
// https://github.com/spf13/cobra/issues/269
func validateCmdName(source string) string {
	i := 0
	l := len(source)
	// The output is initialized on demand, then first dash or underscore
	// occurs.
	var output string

	for i < l {
		if source[i] == '-' || source[i] == '_' {
			if output == "" {
				output = source[:i]
			}

			// If it's last rune and it's dash or underscore,
			// don't add it output and break the loop.
			if i == l-1 {
				break
			}

			// If next character is dash or underscore,
			// just skip the current character.
			if source[i+1] == '-' || source[i+1] == '_' {
				i++
				continue
			}

			// If the current character is dash or underscore,
			// upper next letter and add to output.
			output += string(unicode.ToUpper(rune(source[i+1])))
			// We know, what source[i] is dash or underscore and source[i+1] is
			// uppered character, so make i = i+2.
			i += 2
			continue
		}

		// If the current character isn't dash or underscore,
		// just add it.
		if output != "" {
			output += string(source[i])
		}
		i++
	}

	if output == "" {
		return source // source is initially valid name.
	}
	return output
}
