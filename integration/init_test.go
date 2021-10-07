package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/paketo-buildpacks/occam"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"

	. "github.com/onsi/gomega"
)

var (
	goModBOMBuildpack        string
	offlineGoModBOMBuildpack string
	goDistBuildpack          string
	offlineGoDistBuildpack   string
	goBuildBuildpack         string

	root string

	config struct {
		Buildpack struct {
			ID   string
			Name string
		}
	}

	integrationjson struct {
		GoDist  string `json:"go-dist"`
		GoBuild string `json:"go-build"`
	}
)

func TestIntegration(t *testing.T) {
	Expect := NewWithT(t).Expect

	var err error
	root, err = filepath.Abs("./..")
	Expect(err).ToNot(HaveOccurred())

	file, err := os.Open("../buildpack.toml")
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	_, err = toml.DecodeReader(file, &config)
	Expect(err).NotTo(HaveOccurred())

	file, err = os.Open("../integration.json")
	Expect(err).NotTo(HaveOccurred())

	Expect(json.NewDecoder(file).Decode(&integrationjson)).To(Succeed())
	Expect(file.Close()).To(Succeed())

	buildpackStore := occam.NewBuildpackStore()

	// goModBOMBuildpack, err = buildpackStore.Get.
	// 	WithVersion("1.2.3").
	// 	Execute(root)
	// Expect(err).NotTo(HaveOccurred())

	offlineGoModBOMBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		WithVersion("1.2.3").
		Execute(root)
	Expect(err).NotTo(HaveOccurred())

	goDistBuildpack, err = buildpackStore.Get.
		Execute(integrationjson.GoDist)
	Expect(err).NotTo(HaveOccurred())

	offlineGoDistBuildpack, err = buildpackStore.Get.
		WithOfflineDependencies().
		Execute(integrationjson.GoDist)
	Expect(err).NotTo(HaveOccurred())

	goBuildBuildpack, err = buildpackStore.Get.
		Execute(integrationjson.GoBuild)
	Expect(err).NotTo(HaveOccurred())

	SetDefaultEventuallyTimeout(5 * time.Second)

	suite := spec.New("Integration", spec.Report(report.Terminal{}), spec.Parallel())
	suite("GoMod", testGoMod)
	// suite("Offline", testOffline)
	// suite("PackageLockHashes", testPackageLockHashes)
	// suite("Vendored", testVendored)
	suite.Run(t)
}
