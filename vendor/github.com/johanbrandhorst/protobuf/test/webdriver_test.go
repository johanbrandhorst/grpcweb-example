package webdriver_test

import (
	"fmt"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	"github.com/sclevine/agouti/api"
	. "github.com/sclevine/agouti/matchers"

	"github.com/johanbrandhorst/protobuf/test/shared"
)

var _ = Describe("gRPC-Web Unit Tests", func() {
	if os.Getenv("GOPHERJS_SERVER_ADDR") != "" {
		if os.Getenv("CHROMEDRIVER_ADDR") != "" {
			browserTest("Google Chrome", os.Getenv("GOPHERJS_SERVER_ADDR"), func(opts ...agouti.Option) (*agouti.Page, error) {
				return agouti.NewPage(fmt.Sprintf("http://%s", os.Getenv("CHROMEDRIVER_ADDR")),
					agouti.Desired(agouti.Capabilities{
						"loggingPrefs": map[string]string{
							"browser": "INFO",
						},
					}))
			})
		}

		if os.Getenv("SELENIUM_ADDR") != "" {
			browserTest("Mozilla Firefox", os.Getenv("GOPHERJS_SERVER_ADDR"), func(opts ...agouti.Option) (*agouti.Page, error) {
				return agouti.NewPage(fmt.Sprintf("http://%s/wd/hub", os.Getenv("SELENIUM_ADDR")),
					agouti.Desired(agouti.Capabilities{
						"loggingPrefs": map[string]string{
							"browser": "INFO",
						},
						"acceptInsecureCerts": true,
					}),
					agouti.Browser("firefox"),
				)
			})
		}
	} else {
		browserTest("ChromeDriver", "localhost"+shared.GopherJSServer, chromeDriver.NewPage)
	}
})

type pageFunc func(...agouti.Option) (*agouti.Page, error)

func browserTest(browserName, address string, newPage pageFunc) {
	var page *agouti.Page

	BeforeEach(func() {
		var err error
		page, err = newPage()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		Expect(page.Destroy()).NotTo(HaveOccurred())
	})

	Context(fmt.Sprintf("when testing %s", browserName), func() {
		It("should not find any errors", func() {
			By("Loading the test page", func() {
				Expect(page.Navigate("https://" + address)).NotTo(HaveOccurred())
			})

			By("Finding the number of failures", func() {
				Eventually(page.FirstByClass("failed"), 2).Should(BeFound())
				Eventually(page.FindByID("qunit-testresult").FindByClass("failed"), 2).Should(BeFound())
				numFailures, err := page.FindByID("qunit-testresult").FindByClass("failed").Text()
				Expect(err).NotTo(HaveOccurred())
				if numFailures == "0" {
					return
				}

				logs, err := page.ReadAllLogs("browser")
				Expect(err).NotTo(HaveOccurred())
				fmt.Fprintln(GinkgoWriter, "Console output ------------------------------------")
				for _, log := range logs {
					fmt.Fprintf(GinkgoWriter, "[%s][%s]\t%s\n", log.Time.Format("15:04:05.000"), log.Level, log.Message)
				}
				fmt.Fprintln(GinkgoWriter, "Console output ------------------------------------")

				// We have at least one failure - lets compile an error message
				Eventually(page.AllByXPath("//li[contains(@id, 'qunit-test-output') and @class='fail']")).Should(BeFound())
				failures := page.AllByXPath("//li[contains(@id, 'qunit-test-output') and @class='fail']")
				elements, err := failures.Elements()
				Expect(err).NotTo(HaveOccurred())
				var errMsgs []string
				for _, element := range elements {
					// Get module name
					msg, err := element.GetElement(api.Selector{
						Using: "css selector",
						Value: ".module-name",
					})
					Expect(err).NotTo(HaveOccurred())
					modName, err := msg.GetText()
					Expect(err).NotTo(HaveOccurred())
					// Get test name
					msg, err = element.GetElement(api.Selector{
						Using: "css selector",
						Value: ".test-name",
					})
					Expect(err).NotTo(HaveOccurred())
					testName, err := msg.GetText()
					Expect(err).NotTo(HaveOccurred())
					// Get specific fail node
					fails, err := element.GetElements(api.Selector{
						Using: "css selector",
						Value: ".fail",
					})
					Expect(err).NotTo(HaveOccurred())
					var errSums []string
					for _, fail := range fails {
						// Get error summary
						msg, err := fail.GetElement(api.Selector{
							Using: "css selector",
							Value: ".test-message",
						})
						errSum, err := msg.GetText()
						Expect(err).NotTo(HaveOccurred())

						// Get diff
						expected, err := fail.GetElements(api.Selector{
							Using: "css selector",
							Value: "del",
						})
						Expect(err).NotTo(HaveOccurred())
						var expectedText string
						if len(expected) > 0 {
							expectedText, err = expected[0].GetText()
							Expect(err).NotTo(HaveOccurred())
						}
						actual, err := fail.GetElements(api.Selector{
							Using: "css selector",
							Value: "ins",
						})
						Expect(err).NotTo(HaveOccurred())
						var actualText string
						if len(actual) > 0 {
							actualText, err = actual[0].GetText()
							Expect(err).NotTo(HaveOccurred())
						}
						if expectedText != "" && actualText != "" {
							errSum = fmt.Sprintf(
								"%s\n\t\t\tExpected: %s\n\t\t\tActual: %s",
								errSum,
								strings.TrimSuffix(expectedText, " "),
								strings.TrimSuffix(actualText, " "),
							)
						}

						errSums = append(errSums, errSum)
					}

					errMsgs = append(errMsgs, fmt.Sprintf("%s:\n\t%s:\n\t\t%s", modName, testName, strings.Join(errSums, "\n\t\t")))
				}

				// Prints each error
				Fail(strings.Join(errMsgs, "\n-----------------------------------\n"))
			})
		})
	})
}
