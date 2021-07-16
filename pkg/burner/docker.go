package burner

import (
	"strings"

	backend "github.com/mudler/luet/pkg/compiler/backend"
)

func DockerImages() []string {
	str, err := runO("docker images --format='{{.Repository}}:{{.Tag}}'")
	if err != nil {
		return []string{}
	}
	return strings.Split(str, "\n")
}

func DockerExtract(image, dst string) error {
	docker := backend.NewSimpleDockerBackend()
	return docker.ExtractRootfs(backend.Options{ImageName: image, Destination: dst}, true)
}
