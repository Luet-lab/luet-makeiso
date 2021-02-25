package schema

import (
	"bytes"
	"io"
	"net/http"

	"github.com/twpayne/go-vfs"
	"gopkg.in/yaml.v2"
)

type SystemSpec struct {
	Initramfs   Initramfs  `yaml:"initramfs"`
	Label       string     `yaml:"label"`
	Packages    Packages   `yaml:"packages"`
	Luet        Luet       `yaml:"luet"`
	Repository  Repository `yaml:"repository"`
	Overlay     bool       `yaml:"overlay"`
	ImagePrefix string     `yaml:"image_prefix"`
	Date        bool       `yaml:"image_date"`
}

type Luet struct {
	Config string `yaml:"config"`
}
type Repository struct {
	Initramfs []string `yaml:"initramfs"`
	Packages  []string `yaml:"packages"`
}

type Packages struct {
	KeepLuetDB bool     `yaml:"keep_luet_db"`
	Rootfs     []string `yaml:"rootfs"`
	Initramfs  []string `yaml:"initramfs"`
	IsoImage   []string `yaml:"isoimage"`
	UEFI       []string `yaml:"uefi"`
}

type Initramfs struct {
	KernelFile string `yaml:"kernel_file"`
	RootfsFile string `yaml:"rootfs_file"`
}

// LoadFromFile loads a yip config from a YAML file
func LoadFromFile(s string, fs vfs.FS) (*SystemSpec, error) {
	yamlFile, err := fs.ReadFile(s)
	if err != nil {
		return nil, err
	}

	return LoadFromYaml(yamlFile)
}

// LoadFromYaml loads a yip config from bytes
func LoadFromYaml(b []byte) (*SystemSpec, error) {

	var yamlConfig SystemSpec
	err := yaml.Unmarshal(b, &yamlConfig)
	if err != nil {
		return nil, err
	}

	return &yamlConfig, nil
}

// LoadFromUrl loads a yip config from a url
func LoadFromUrl(s string) (*SystemSpec, error) {
	resp, err := http.Get(s)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := bytes.NewBuffer([]byte{})
	_, err = io.Copy(buf, resp.Body)

	return LoadFromYaml(buf.Bytes())
}
