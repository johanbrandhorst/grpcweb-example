package webdriver_test

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/sclevine/agouti"
)

func TestWebdriver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Webdriver Suite")
}

var (
	testServer   *gexec.Session
	chromeDriver = agouti.ChromeDriver(
		agouti.Desired(agouti.Capabilities{
			"loggingPrefs": map[string]string{
				"browser": "INFO",
			},
			"browserName": "chrome",
		}),
		// Unfortunately headless doesn't seem to work quite yet,
		// seems lock up loading the page.
		// (tried Google Chrome 59.0.3071.115)
		// https://developers.google.com/web/updates/2017/04/headless-chrome#drivers
		/*agouti.ChromeOptions(
			"args", []string{
				"--headless",
				"--disable-gpu",
			},
		),
		agouti.ChromeOptions(
			"binary", "/usr/bin/google-chrome-stable",
		),*/
	)
	seleniumDriver = agouti.Selenium(
		agouti.Browser("firefox"),
		agouti.Desired(agouti.NewCapabilities("acceptInsecureCerts")),
	)
)

var _ = BeforeSuite(func() {
	var binPath string
	By("Building the server", func() {
		var err error
		binPath, err = gexec.Build("./server/main.go")
		Expect(err).NotTo(HaveOccurred())
	})

	By("Running the server", func() {
		var err error
		testServer, err = gexec.Start(exec.Command(binPath), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
	})

	By("Starting the WebDrivers", func() {
		if os.Getenv("GOPHERJS_SERVER_ADDR") == "" {
			Expect(chromeDriver.Start()).NotTo(HaveOccurred())
			//Expect(seleniumDriver.Start()).NotTo(HaveOccurred())
		}
	})
})

var _ = AfterSuite(func() {
	By("Stopping the WebDrivers", func() {
		if os.Getenv("GOPHERJS_SERVER_ADDR") == "" {
			Expect(chromeDriver.Stop()).NotTo(HaveOccurred())
			//Expect(seleniumDriver.Stop()).NotTo(HaveOccurred())
		}
	})

	By("Stopping the server", func() {
		testServer.Terminate()
		testServer.Wait()
		Expect(testServer).To(gexec.Exit())
	})

	By("Cleaning up built artifacts", func() {
		gexec.CleanupBuildArtifacts()
	})
})
