package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mudler/luet-makeiso/pkg/burner"
	"github.com/mudler/luet-makeiso/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/twpayne/go-vfs"
)

const (
	CLIVersion = "0.4.0"
)

// Build time and commit information.
//
// ⚠️ WARNING: should only be set by "-ldflags".
var (
	BuildTime   string
	BuildCommit string
)

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
	Use:     "luet-makeiso",
	Short:   "generate iso images with luet",
	Version: fmt.Sprintf("%s-g%s %s", CLIVersion, BuildCommit, BuildTime),
	Long: `It reads iso spec to generate ISO files from luet repositories or trees.
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Error("One argument (spec) required")
			os.Exit(1)
		}

		localPath, _ := cmd.Flags().GetString("local")
		image, _ := cmd.Flags().GetString("image")
		output, _ := cmd.Flags().GetString("output")
		skipCleanup, _ := cmd.Flags().GetBool("skip-cleanup")

		if localPath != "" && !filepath.IsAbs(localPath) {
			var err error
			localPath, err = filepath.Abs(localPath)
			checkErr(err)
		}

		for _, a := range args {
			spec, err := schema.LoadFromFile(a, vfs.OSFS)
			checkErr(err)

			if image != "" {
				spec.RootfsImage = image
			}
			if output != "" {
				spec.ImageName = output
				spec.Date = false
				spec.ImagePrefix = ""
			}

			if localPath != "" {
				spec.Luet.Repositories = append(spec.Luet.Repositories, schema.NewLocalRepo("local", localPath))
			}
			checkErr(burner.Burn(spec, vfs.OSFS, !skipCleanup))
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	checkErr(err)
}

func init() {
	rootCmd.Flags().StringP("local", "l", "", "A path to a local luet repository to use during iso build")
	rootCmd.Flags().StringP("image", "i", "", "An image reference to use as a rootfs for the ISO")
	rootCmd.Flags().StringP("output", "o", "", "Name of the output ISO file (overrides yaml config)")
	rootCmd.Flags().Bool("skip-cleanup", false, "Avoid to delete the staging directories.")
}
