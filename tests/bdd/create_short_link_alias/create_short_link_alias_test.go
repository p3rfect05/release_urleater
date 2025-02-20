package create_short_link_alias

import (
	"context"
	"flag"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/tebeka/selenium"
	"os"
	"strings"
	"testing"
	"time"
)

var opts = godog.Options{
	Format: "pretty",
	Paths:  []string{"features"},
	Output: colors.Colored(os.Stdout),
}

const (
	// These paths will be different on your system.
	port        = 4444
	browserName = "firefox"
	site        = "http://localhost:8080"
)

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opts)
}

func TestCreateShortLink(t *testing.T) {
	o := opts

	o.TestingT = t

	status := godog.TestSuite{
		Name:                "godogs",
		Options:             &o,
		ScenarioInitializer: InitializeScenario,
	}.Run()

	if status != 0 {
		t.Fatalf("zero status code expected, %d received", status)
	}
}

func configureDriver(ctx context.Context, svc *godog.Scenario) (context.Context, error) {
	caps := selenium.Capabilities{"browserName": browserName}

	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d", port))

	if err != nil {
		panic(err)
	}

	ctx = context.WithValue(ctx, "driver", wd)
	time.Sleep(time.Second)
	return ctx, nil
}

func disableDriver(ctx context.Context, svc *godog.Scenario, err error) (context.Context, error) {
	driver := ctx.Value("driver").(selenium.WebDriver)
	time.Sleep(5 * time.Second)
	err = driver.Quit()

	if err != nil {
		panic(err)
	}

	return ctx, nil
}

func iHaveLoggedIn(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	err := driver.Get(site + "/login")
	if err != nil {
		panic(err)
	}

	elem1, err := driver.FindElement(selenium.ByID, "email")
	if err != nil {
		panic(err)
	}

	err = elem1.SendKeys("test@example.com")
	if err != nil {
		panic(err)
	}

	elem2, err := driver.FindElement(selenium.ByID, "password")
	if err != nil {
		panic(err)
	}

	err = elem2.SendKeys("12341234")
	if err != nil {
		panic(err)
	}

	elem3, err := driver.FindElement(selenium.ByCSSSelector, "body > div > form > button")
	if err != nil {
		panic(err)
	}

	err = elem3.Click()
	if err != nil {
		panic(err)
	}

	return nil
}

func iClickedOnCreateShortLinkTab(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	err := driver.Get(site + "/create_link")
	if err != nil {
		panic(err)
	}

	return nil
}

func iTypeLongUrl(ctx context.Context, longUrl string) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem1, err := driver.FindElement(selenium.ByID, "longUrl")
	if err != nil {
		panic(err)
	}

	err = elem1.SendKeys(longUrl)
	if err != nil {
		panic(err)
	}

	return nil
}

func iTypeAlias(ctx context.Context, alias string) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem1, err := driver.FindElement(selenium.ByID, "customPath")
	if err != nil {
		panic(err)
	}

	err = elem1.SendKeys(alias)
	if err != nil {
		panic(err)
	}

	return nil
}

func iChooseGeneration(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem, err := driver.FindElement(selenium.ByID, "customAlias")

	if err != nil {
		panic(err)
	}

	err = elem.Click()
	if err != nil {
		panic(err)
	}

	return nil
}

func iClickOnCreateButton(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem, err := driver.FindElement(selenium.ByID, "generateBtn")

	if err != nil {
		panic(err)
	}

	err = elem.Click()
	if err != nil {
		panic(err)
	}

	return nil
}

func iGetNewlyGeneratedLink(ctx context.Context, alias string) error {
	driver := ctx.Value("driver").(selenium.WebDriver)
	err := driver.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
		elem, err := driver.FindElement(selenium.ByID, "qr_code")
		if err != nil {
			panic(err)
		}

		return elem.IsDisplayed()
	}, 10*time.Second, 500*time.Millisecond)
	if err != nil {
		panic(err)
	}
	elem, err := driver.FindElement(selenium.ByID, "qr_code")
	if err != nil {
		panic(err)
	}

	text, err := elem.GetAttribute("title")
	if err != nil {
		panic(err)
	}

	if !strings.Contains(text, alias) {
		panic(fmt.Sprintf("no link generated %s", text))
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(configureDriver)
	ctx.Given(`^I have logged in$`, iHaveLoggedIn)
	ctx.Given(`^I clicked on create short link tab$`, iClickedOnCreateShortLinkTab)
	ctx.When(`^I type url (.*)$`, iTypeLongUrl)
	ctx.When(`^I choose generation$`, iChooseGeneration)
	ctx.When(`^I type alias (.*)$`, iTypeAlias)
	ctx.When(`^I click on create button$`, iClickOnCreateButton)
	ctx.Then(`^I get newly generated Link with alias (.*)$`, iGetNewlyGeneratedLink)
	ctx.After(disableDriver)
}
