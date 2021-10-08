package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/occam"
	"github.com/paketo-buildpacks/packit/fs"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
	. "github.com/paketo-buildpacks/occam/matchers"
)

func testMultipleMain(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		pack   occam.Pack
		docker occam.Docker
	)

	it.Before(func() {
		pack = occam.NewPack()
		docker = occam.NewDocker()
	})

	context("when the buildpack is run with pack build", func() {
		var (
			image     occam.Image
			container occam.Container
			name      string
			source    string
		)

		it.Before(func() {
			var err error
			name, err = occam.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(docker.Image.Remove.Execute(image.ID)).To(Succeed())
			Expect(docker.Volume.Remove.Execute(occam.CacheVolumeNames(name))).To(Succeed())
			Expect(os.RemoveAll(source)).To(Succeed())
		})

		context("building a go mod app from multiple main files is pack built", func() {
			it.After(func() {
				Expect(docker.Container.Remove.Execute(container.ID)).To(Succeed())
			})

			it("builds, logs and runs one of the main entrypoints correctly", func() {
				var err error

				source, err = occam.Source(filepath.Join("testdata", "multimain"))
				Expect(err).ToNot(HaveOccurred())

				err = replaceGitFileWithSubmoduleDir(source)
				Expect(err).ToNot(HaveOccurred())

				var logs fmt.Stringer
				image, logs, err = pack.WithNoColor().Build.
					WithPullPolicy("never").
					WithBuildpacks(
						goDistBuildpack,
						offlineGoModBOMBuildpack, // TODO: Use online buildpack here once we resolve the packaging issue of cyclonedx-gomod (it needs to have a bin directory)
						// goModBOMBuildpack,
						goBuildBuildpack,
					).
					Execute(name, source)
				Expect(err).ToNot(HaveOccurred(), logs.String)

				container, err = docker.Container.Run.
					WithPublish("8080").
					Execute(image.ID)
				Expect(err).NotTo(HaveOccurred())

				Eventually(container).Should(BeAvailable())
				Eventually(container).Should(Serve(ContainSubstring("Random UUID")).OnPort(8080))
				Expect(image.Labels["io.buildpacks.build.metadata"]).To(ContainSubstring(`"name":"github.com/robdimsdale/multimain"`))
				Expect(image.Labels["io.buildpacks.build.metadata"]).To(ContainSubstring(`"name":"github.com/robdimsdale/uuid"`))
				Expect(image.Labels["io.buildpacks.build.metadata"]).To(ContainSubstring(`"name":"github.com/google/uuid"`))

				Expect(image.Labels["io.buildpacks.build.metadata"]).NotTo(ContainSubstring(`"name":"gopkg.in/yaml.v2"`))
			})
		})
	})
}

func replaceGitFileWithSubmoduleDir(source string) error {
	gitfile := filepath.Join(source, ".git")

	err := os.Remove(gitfile)
	if err != nil {
		return err
	}

	gitSubmoduleDir := filepath.Join(".", "../.git/modules/integration/testdata/simple-golang-uuid")

	return fs.Copy(gitSubmoduleDir, gitfile)
}
