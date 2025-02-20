package login

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
	Paths:  []string{"./features/login.feature"},
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

func TestLogin(t *testing.T) {
	o := opts

	o.TestingT = t

	status := godog.TestSuite{
		Name:                "Login",
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

func iHaveNavigatedToMainPageAndNotLoggedIn(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	err := driver.Get(site + "/login")
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second)

	elem, err := driver.FindElement(selenium.ByCSSSelector, "body > div > h3")
	if err != nil {
		panic(err)
	}

	text, err := elem.Text()
	if err != nil {
		panic(err)
	}

	if text != "Вход" {
		panic(err)
	}

	return nil
}

func iTypeEmail(ctx context.Context, email string) error {
	driver := ctx.Value("driver").(selenium.WebDriver)
	elem, err := driver.FindElement(selenium.ByID, "email")
	if err != nil {
		panic(err)
	}

	err = elem.SendKeys(email)
	if err != nil {
		panic(err)
	}

	return nil
}

func iTypePassword(ctx context.Context, password string) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem, err := driver.FindElement(selenium.ByID, "password")
	if err != nil {
		panic(err)
	}

	err = elem.SendKeys(password)
	if err != nil {
		panic(err)
	}

	return nil
}

func iClickedLogin(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	err := driver.WaitWithTimeoutAndInterval(func(driver selenium.WebDriver) (bool, error) {
		elem, err := driver.FindElement(selenium.ByCSSSelector, "body > div > form > button")
		if err != nil {
			panic(err)
		}
		return elem.IsDisplayed()
	}, 10*time.Second, 500*time.Millisecond)

	if err != nil {
		panic(err)
	}

	elem, err := driver.FindElement(selenium.ByCSSSelector, "body > div > form > button")

	if err != nil {
		panic(err)
	}

	err = elem.Click()
	if err != nil {
		panic(err)
	}

	return nil
}

func iShouldSeeTheMainPageAndBeLoggedIn(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem, err := driver.FindElement(selenium.ByCSSSelector, "#user_links > div > div > h1")
	if err != nil {
		panic(err)
	}

	text, err := elem.Text()
	if err != nil {
		panic(err)
	}

	if !strings.Contains(text, "Главную") {
		panic(fmt.Sprintf("we are not on the main page: %s", text))
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(configureDriver)
	ctx.Given(`^I have navigated to main page and not logged in$`, iHaveNavigatedToMainPageAndNotLoggedIn)
	ctx.When(`^I type email (.*)$`, iTypeEmail)
	ctx.When(`^I type password (.*)$`, iTypePassword)
	ctx.When(`^I clicked Login$`, iClickedLogin)
	ctx.Then(`^I should see the main page and be logged in$`, iShouldSeeTheMainPageAndBeLoggedIn)
	ctx.After(disableDriver)

}
func iTypeEmailEmail(ctx context.Context, email string) (context.Context, error) {
	return ctx, godog.ErrPending
}
