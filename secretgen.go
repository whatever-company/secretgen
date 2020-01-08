package secretgen

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"

	"go.mozilla.org/sops/v3/cmd/sops/common"
	"go.mozilla.org/sops/v3/cmd/sops/formats"
	"go.mozilla.org/sops/v3/decrypt"
)

var suffixMap = map[string]string{
	"json": "json",
	"yaml": "yaml",
	"yml":  "yaml",
	"env":  "dotenv",
}

var linesRe = regexp.MustCompile("[\\r\\n]+")

type File struct {
	Key      string   `yaml:"key"`
	File     string   `yaml:"file"`
	TryFiles []string `yaml:"tryFiles"`
}

type Secret struct {
	Name      string   `yaml:"name"`
	Namespace string   `yaml:"namespace"`
	Envs      []string `yaml:"envs"`
	Files     []File   `yaml:"files"`
	Behavior  string   `yaml:"behavior"`
}

type Config struct {
	Secrets []Secret `yaml:"secrets"`
}

type Metadata struct {
	Name        string            `yaml:"name"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type SecretManifest struct {
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Type       string            `yaml:"type"`
	Metadata   Metadata          `yaml:"metadata"`
	Data       map[string]string `yaml:"data"`
}

func decryptData(content []byte, format string, fake bool) ([]byte, error) {
	if fake {
		formatFmt := formats.FormatFromString(format)
		store := common.StoreForFormat(formatFmt)
		tree, err := store.LoadEncryptedFile(content)
		check(err)
		return store.EmitPlainFile(tree.Branches)
	} else {
		return decrypt.Data(content, format)
	}
}

func Generate(c Config, fake bool) []SecretManifest {
	secretManifests := []SecretManifest{}

	for _, secret := range c.Secrets {
		secretData := make(map[string]string)

		for _, fPath := range secret.Envs {
			content, err := ioutil.ReadFile(fPath)
			check(err)
			content, err = decryptData(content, "dotenv", fake)
			check(err)

			handledLine := false
			for i, line := range linesRe.Split(string(content), -1) {
				if len(line) == 0 || strings.HasPrefix(line, "#") {
					continue
				}
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					log.Fatalf("Invalid format on line %d", i)
				}
				secretData[parts[0]] = encode([]byte(parts[1]))
				handledLine = true
			}
			if !handledLine {
				log.Fatalf("env did not contain any secrets %q", content)
			}
		}

		for _, file := range secret.Files {
			if file.File != "" && len(file.TryFiles) > 0 || file.File == "" && len(file.TryFiles) == 0 {
				log.Fatalf("must specify exactly one of file and tryFiles, got file=%q, tryFiles=%q", file.File, file.TryFiles)
			}

			if file.File != "" {
				content, err := ioutil.ReadFile(file.File)
				check(err)
				content, err = decryptData(content, modeForFilename(file.File), fake)
				check(err)
				secretData[file.Key] = encode(content)
			} else {
				handledAFile := false
				for _, tryFile := range file.TryFiles {
					content, err := ioutil.ReadFile(tryFile)
					if err != nil { // skip to next file
						continue
					}
					content, err = decryptData(content, modeForFilename(tryFile), fake)
					check(err)
					secretData[file.Key] = encode(content)
					handledAFile = true
				}
				if !handledAFile {
					log.Fatalf("could not load any of %q", file.TryFiles)
				}
			}
		}

		annotations := map[string]string{"kustomize.config.k8s.io/needs-hash": "true"}
		if secret.Behavior != "" {
			annotations["kustomize.config.k8s.io/behavior"] = secret.Behavior
		}
		manifest := SecretManifest{
			APIVersion: "v1",
			Kind:       "Secret",
			Metadata: Metadata{
				Name:        secret.Name,
				Namespace:   secret.Namespace,
				Annotations: annotations,
			}, Type: "Opaque",
			Data: secretData,
		}
		secretManifests = append(secretManifests, manifest)
	}

	return secretManifests
}

func check(e error) {
	if e != nil {
		log.Panic(e)
	}
}

func modeForFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if strings.HasPrefix(ext, ".") {
		ext = ext[1:]
	}
	mode, ok := suffixMap[ext]
	if !ok {
		return "binary"
	}
	return mode
}

func encode(value []byte) string {
	return base64.StdEncoding.EncodeToString(value)
}
