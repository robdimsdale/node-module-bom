package main

import (
	"os"

	gomodbom "github.com/paketo-buildpacks/go-mod-bom"
	"github.com/paketo-buildpacks/packit"
	"github.com/paketo-buildpacks/packit/cargo"
	"github.com/paketo-buildpacks/packit/chronos"
	"github.com/paketo-buildpacks/packit/pexec"
	"github.com/paketo-buildpacks/packit/postal"
	"github.com/paketo-buildpacks/packit/scribe"
)

func main() {
	goModParser := gomodbom.NewGoModParser()

	packit.Run(
		gomodbom.Detect(goModParser),
		gomodbom.Build(
			postal.NewService(cargo.NewTransport()),
			gomodbom.NewModuleBOM(pexec.NewExecutable("cyclonedx-gomod"), scribe.NewEmitter(os.Stdout)),
			chronos.DefaultClock,
			scribe.NewEmitter(os.Stdout),
		),
	)
}
