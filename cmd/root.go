// Copyright © 2020 Ettore Di Giacinto <mudler@gentoo.org>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/mudler/luet-geniso/pkg/burner"
	"github.com/mudler/luet-geniso/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

const (
	CLIVersion = "0.1"
)

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

var entityFile string

func fail(s string) {
	log.Error(s)
	os.Exit(1)
}
func checkErr(err error) {
	if err != nil {
		fail("fatal error: " + err.Error())
	}
}

func init() {
	switch strings.ToLower(os.Getenv("LOGLEVEL")) {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warning":
		log.SetLevel(log.WarnLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "luet-geniso",
	Short:   "generate iso files with luet",
	Version: fmt.Sprintf("%s-g%s %s", CLIVersion, BuildCommit, BuildTime),
	Long: `It reads iso spec to generate ISO files from luet repositories or trees.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Error("One argument (spec) required")
			os.Exit(1)
		}
		for _, a := range args {
			spec, err := schema.LoadFromFile(a, vfs.OSFS)
			checkErr(err)

			checkErr(burner.Burn(spec, vfs.OSFS))
		}

		//	fs := vfs.OSFS
		//stage, _ := cmd.Flags().GetString("stage")
		//exec, _ := cmd.Flags().GetString("executor")
		//runner := executor.NewExecutor(exec)
		//fromStdin := len(args) == 1 && args[0] == "-"

		//checkErr(runner.Run(stage, vfs.OSFS, stdConsole, args...))
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	checkErr(err)
}

func init() {
	//rootCmd.PersistentFlags().StringP("spec", "e", "default", "Executor which applies the config")
	//rootCmd.PersistentFlags().StringP("spec", "s", "", "Specfile")
}
