package burner

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mudler/luet-makeiso/pkg/schema"
	log "github.com/sirupsen/logrus"
	"github.com/twpayne/go-vfs"
)

func run(cmd string, opts ...func(cmd *exec.Cmd)) (string, error) {
	log.Debugf("running command `%s`", cmd)
	c := exec.Command("sh", "-c", cmd)
	for _, o := range opts {
		o(c)
	}
	c.Env = []string{"LUET_NOLOCK=true", "LUET_YES=true"}
	out, err := c.CombinedOutput()
	if err != nil {
		return string(out), fmt.Errorf("failed to run %s: %v", cmd, err)
	}

	return string(out), err
}

func copyConfig(config, rootfsWanted string, fs vfs.FS, s *schema.SystemSpec) error {
	//	cfg, _ := fs.RawPath(s.Luet.Config)
	repos, err := s.Luet.Repositories.Marshal()
	if err != nil {
		return err
	}
	rootfs, _ := fs.RawPath(rootfsWanted)

	// input, err := ioutil.ReadFile(cfg)
	// if err != nil {
	// 	return err
	// }
	// XXX: This is temporarly needed until https://github.com/mudler/luet/issues/186 is closed
	input := []byte(repos + "\n" +
		`
system:
  rootfs: ` + rootfs + `
  database_path: "/luetdb"
  database_engine: "boltdb"
repos_confdir:
  - ` + rootfs + `/etc/luet/repos.conf.d
` + "\n")
	err = fs.WriteFile(config, input, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LuetInstall(rootfs string, packages []string, repositories []string, keepDB bool, fs vfs.FS, spec *schema.SystemSpec) error {
	cfgFile := filepath.Join(rootfs, "luet.yaml")
	cfgRaw, _ := fs.RawPath(cfgFile)

	if err := copyConfig(cfgFile, rootfs, fs, spec); err != nil {
		return err
	}

	if len(repositories) > 0 {
		out, err := run(fmt.Sprintf("luet install --config %s %s", cfgRaw, strings.Join(repositories, " ")))
		log.Info(out)
		if err != nil {
			return err
		}
	}

	if len(packages) > 0 {
		out, err := run(fmt.Sprintf("luet install --config %s %s", cfgRaw, strings.Join(packages, " ")))
		log.Info(out)
		if err != nil {
			return err
		}
	}

	out, err := run(fmt.Sprintf("luet --config %s cleanup", cfgRaw))
	log.Info(out)
	if err != nil {
		return err
	}

	if keepDB {
		if err := vfs.MkdirAll(fs, filepath.Join(rootfs, "var", "luet"), os.ModePerm); err != nil {
			return err
		}
		if _, err := fs.Stat(filepath.Join(rootfs, "var", "luet", "db")); err == nil {
			fs.RemoveAll(filepath.Join(rootfs, "var", "luet", "db"))
		}
		fs.Rename(filepath.Join(rootfs, "luetdb"), filepath.Join(rootfs, "var", "luet", "db"))
	} else {
		fs.RemoveAll(filepath.Join(rootfs, "luetdb"))
	}
	fs.Remove(cfgFile)
	fs.Remove(filepath.Join(rootfs, "luet", "repos.conf.d"))
	return nil
}
