package nodemodulebom

import (
	"io/ioutil"
	"time"

	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/sbom"
	"github.com/paketo-buildpacks/packit/scribe"
)

//go:generate faux --interface DependencyManager --output fakes/dependency_manager.go
type DependencyManager interface {
	Resolve(path, id, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

func Build(dependencyManager DependencyManager, clock chronos.Clock, logger scribe.Emitter) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		layer, err := context.Layers.Get("node-module-bom")
		if err != nil {
			return packit.BuildResult{}, err
		}

		layer, err = layer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Process("Generating SBOM for directory %s", context.WorkingDir)

		var bom sbom.SBOM
		duration, err := clock.Measure(func() error {
			bom, err = sbom.Generate(context.WorkingDir)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		entries := sbom.NewEntries(bom)
		entries.AddFormat(sbom.CycloneDXFormat)
		entries.AddFormat(sbom.SyftFormat)
		entries.AddFormat(sbom.SPDXFormat)

		layer.SBOM = entries
		layer.Launch = true

		b, err := ioutil.ReadAll(entries.GetContent(sbom.CycloneDXFormat))
		if err != nil {
			panic(err)
		}

		logger.Detail("bom content:\n%s", string(b))

		return packit.BuildResult{
			Layers: []packit.Layer{
				layer,
			},
			Build: packit.BuildMetadata{},
			Launch: packit.LaunchMetadata{
				SBOM: entries,
			},
		}, nil
	}
}
