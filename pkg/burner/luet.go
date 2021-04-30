package burner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mudler/luet-makeiso/pkg/schema"
	"github.com/twpayne/go-vfs"
)

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
		if err := run(fmt.Sprintf("luet install --no-spinner --config %s %s", cfgRaw, strings.Join(repositories, " "))); err != nil {
			return err
		}
	}

	if len(packages) > 0 {
		if err := run(fmt.Sprintf("luet install --no-spinner --config %s %s", cfgRaw, strings.Join(packages, " "))); err != nil {
			return err
		}
	}

	if err := run(fmt.Sprintf("luet --config %s cleanup", cfgRaw)); err != nil {
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
