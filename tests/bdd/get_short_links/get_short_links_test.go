package get_short_links

import (
	"context"
	"flag"
	"fmt"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/tebeka/selenium"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type driver struct{}

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

func TestGetShortLinks(t *testing.T) {
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

	ctx = context.WithValue(ctx, driver{}, wd)
	time.Sleep(time.Second)
	return ctx, nil
}

func disableDriver(ctx context.Context, svc *godog.Scenario, err error) (context.Context, error) {
	if err != nil {
		panic(err)
	}

	driver := ctx.Value(driver{}).(selenium.WebDriver)
	time.Sleep(5 * time.Second)
	err = driver.Quit()

	if err != nil {
		panic(err)
	}

	return ctx, nil
}

func iHaveLoggedIn(ctx context.Context) error {
	driver := ctx.Value(driver{}).(selenium.WebDriver)

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

func iHaveMultipleLinksCreated(ctx context.Context) error {
	driver := ctx.Value(driver{}).(selenium.WebDriver)

	err := driver.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
		elem, err := driver.FindElement(selenium.ByID, "display_number_of_links")
		if err != nil {
			panic(err)
		}
		text, err := elem.Text()
		if err != nil {
			panic(err)
		}
		splittedValues := strings.Split(text, " ")
		activeLinks, err := strconv.Atoi(splittedValues[len(splittedValues)-1])
		if err != nil {
			return false, nil
		}
		return activeLinks > 0, nil
	}, 10*time.Second, 1*time.Second)

	if err != nil {
		panic(err)
	}

	return nil
}

func iClickedOnMyLinksLinkTab(ctx context.Context) error {
	driver := ctx.Value(driver{}).(selenium.WebDriver)

	err := driver.Get(site + "/my_links")
	if err != nil {
		panic(err)
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(configureDriver)
	ctx.Given(`^I have logged in$`, iHaveLoggedIn)
	ctx.Given(`^I clicked on my links tab$`, iClickedOnMyLinksLinkTab)
	ctx.Then(`^I have multiple links created$`, iHaveMultipleLinksCreated)
	ctx.After(disableDriver)
}
