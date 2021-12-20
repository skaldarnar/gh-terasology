package cmd

import (
	"fmt"
	"os"
	"path"
	"strings"

	git "github.com/cli/cli/v2/git"
	"github.com/spf13/cobra"
)

type WorkspaceOptions struct {
}

func NewWorkspaceCommand() *cobra.Command {
	//opts := &WorkspaceOptions{}

	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			toplevelDir, err := TerasologyToplevelDir(".")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Println(toplevelDir)
		},
	}
	return cmd
}

func init() {
	rootCmd.AddCommand(NewWorkspaceCommand())
}

// --------------------------------------------------------------------------------------------------------------------

// Look for the toplevel directory (root directory) of a Terasology checkout.
//
// A Terasology checkout is determined by a heuristic by looking for a `settings.gradle` file
// in the toplevel directory of a git repository and comparing the `rootProject.name` against
// `'Terasology'`.
func TerasologyToplevelDir(dir string) (string, error) {
	err := os.Chdir(dir)
	if err != nil {
		return "", err
	}

	toplevel, err := git.ToplevelDir()
	if err != nil {
		return "", err
	}

	found, err := isTerasologyToplevel(toplevel)
	if err != nil {
		return "", err
	} else if found {
		return toplevel, nil
	}

	return TerasologyToplevelDir(path.Dir(toplevel))
}

func isTerasologyToplevel(dir string) (bool, error) {
	// TODO: improve check for Terasology root directory
	//				- line-based file reader to stop on first occurrence of `rootProject.name`
	//			  - other heuristic?
	settingsPath := path.Join(dir, "settings.gradle")
	if _, err := os.Stat(settingsPath); err == nil {
		dat, err := os.ReadFile(settingsPath)
		if err != nil {
			return false, err
		}
		if strings.Contains(string(dat), `rootProject.name = 'Terasology'`) {
			return true, nil
		}
	}
	return false, nil
}
