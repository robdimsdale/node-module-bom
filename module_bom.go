package gomodbom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface Executable --output fakes/executable.go
type Executable interface {
	Execute(execution pexec.Execution) error
}

type ModuleBOM struct {
	executable Executable
	logger     scribe.Emitter
}

func NewModuleBOM(executable Executable, logger scribe.Emitter) ModuleBOM {
	return ModuleBOM{
		executable: executable,
		logger:     logger,
	}
}

func (m ModuleBOM) Generate(workingDir, target string) ([]packit.BOMEntry, error) {
	buffer := bytes.NewBuffer(nil)
	args := []string{
		"app",
		"-json",
		"-files",
		"-licenses",
		"-main", target,
		"-output", "bom.json",
	}
	m.logger.Subprocess("Running 'cyclonedx-gomod %s'", strings.Join(args, " "))
	err := m.executable.Execute(pexec.Execution{
		Args:   args,
		Dir:    workingDir,
		Stdout: buffer,
		Stderr: buffer,
	})

	if err != nil {
		m.logger.Detail(buffer.String())
		return nil, fmt.Errorf("failed to run cyclonedx-gomod: %w", err)
	}

	file, err := os.Open(filepath.Join(workingDir, "bom.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to open bom.json: %w", err)
	}
	defer file.Close()

	// TODO: figure out why the licenses field isn't present. Probably needs 'evidences'.
	var bom struct {
		Components []struct {
			Name    string `json:"name"`
			PURL    string `json:"purl"`
			Version string `json:"version"`
			Hashes  []struct {
				Algorithm string `json:"alg"`
				Content   string `json:"content"`
			} `json:"hashes"`
			Licenses []struct {
				License struct {
					ID string `json:"id"`
				} `json:"license"`
			} `json:"licenses"`
		} `json:"components"`
	}

	err = json.NewDecoder(file).Decode(&bom)
	if err != nil {
		return nil, fmt.Errorf("failed to decode bom.json: %w", err)
	}

	var entries []packit.BOMEntry
	for _, entry := range bom.Components {
		packitEntry := packit.BOMEntry{
			Name: entry.Name,
			Metadata: packit.BOMMetadata{
				Version: entry.Version,
				PURL:    entry.PURL,
			},
		}

		if len(entry.Hashes) > 0 {
			algorithm, err := packit.GetBOMChecksumAlgorithm(entry.Hashes[0].Algorithm)
			if err != nil {
				return nil, err
			}
			packitEntry.Metadata.Checksum = packit.BOMChecksum{
				Algorithm: algorithm,
				Hash:      entry.Hashes[0].Content,
			}
		}

		var licenses []string
		for _, license := range entry.Licenses {
			licenses = append(licenses, license.License.ID)
		}
		packitEntry.Metadata.Licenses = licenses
		entries = append(entries, packitEntry)
	}

	err = os.Remove(filepath.Join(workingDir, "bom.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to remove bom.json: %w", err)
	}

	return entries, nil
}
