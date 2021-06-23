package schema

import (
	"time"

	"github.com/twpayne/go-vfs"
	"gopkg.in/yaml.v2"
)

type SystemSpec struct {
	Initramfs   Initramfs  `yaml:"initramfs"`
	Label       string     `yaml:"label"`
	Packages    Packages   `yaml:"packages"`
	Luet        Luet       `yaml:"luet"`
	Repository  Repository `yaml:"repository"`
	ImagePrefix string     `yaml:"image_prefix"`
	Date        bool       `yaml:"image_date"`
	ImageName   string     `yaml:"image_name"`
	Arch        string     `yaml:"arch"`
	UEFIImage   string     `yaml:"uefi_img"`
}

type Luet struct {
	Repositories Repositories `yaml:"repositories"`
}

type LuetRepository struct {
	Name     string   `yaml:"name"`
	Enable   bool     `yaml:"enable"`
	Urls     []string `yaml:"urls"`
	Type     string   `yaml:"type"`
	Priority int      `yaml:"priority"`
}

type Repositories []*LuetRepository

func (r Repositories) Marshal() (string, error) {
	b, err := yaml.Marshal(&Luet{Repositories: r})

	return string(b), err
}

func genRepo(name, url, t string) *LuetRepository {
	return &LuetRepository{Name: name, Enable: true, Urls: []string{url}, Type: t}
}

func NewDockerRepo(name, url string) *LuetRepository {
	return genRepo(name, url, "docker")
}

func NewHTTPRepo(name, url string) *LuetRepository {
	return genRepo(name, url, "http")
}

func NewLocalRepo(name, path string) *LuetRepository {
	return genRepo(name, path, "disk")
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

func (s *SystemSpec) ISOName() (imageName string) {
	if s.ImageName != "" {
		imageName = s.ImageName
	}
	if s.ImagePrefix != "" {
		imageName = s.ImagePrefix + imageName
	}
	if s.Date {
		currentTime := time.Now()
		imageName = imageName + currentTime.Format("20060102")
	}
	if imageName == "" {
		imageName = "dev"
	}
	imageName = imageName + ".iso"
	return
}

// LoadFromFile loads a yip config from a YAML file
func LoadFromFile(s string, fs vfs.FS) (*SystemSpec, error) {
	yamlFile, err := fs.ReadFile(s)
	if err != nil {
		return nil, err
	}

	return LoadFromYaml(yamlFile)
}

// LoadFromYaml loads a config from bytes
func LoadFromYaml(b []byte) (*SystemSpec, error) {

	var yamlConfig SystemSpec
	err := yaml.Unmarshal(b, &yamlConfig)
	if err != nil {
		return nil, err
	}

	return &yamlConfig, nil
}
