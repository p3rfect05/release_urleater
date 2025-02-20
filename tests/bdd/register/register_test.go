package register

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

func TestRegister(t *testing.T) {
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

func IHaveNavigatedToMainPageAndNotLoggedIn(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	err := driver.Get(site + "/register")
	if err != nil {
		panic(err)
	}

	elem, err := driver.FindElement(selenium.ByCSSSelector, "body > div > h3")
	if err != nil {
		panic(err)
	}

	text, err := elem.Text()
	if err != nil {
		panic(err)
	}

	if text != "Регистрация" {
		panic(err)
	}

	return nil
}

func typeEmail(ctx context.Context, email string) error {
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

func typePassword(ctx context.Context, password string) error {
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

func typeAgainPassword(ctx context.Context, password string) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem, err := driver.FindElement(selenium.ByID, "confirm-password")
	if err != nil {
		panic(err)
	}

	err = elem.SendKeys(password)
	if err != nil {
		panic(err)
	}

	return nil
}

func clickRegister(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem, err := driver.FindElement(selenium.ByCSSSelector, "#registrationForm > button")
	if err != nil {
		panic(err)
	}

	err = elem.Click()
	if err != nil {
		panic(err)
	}

	return nil
}

func checkIfOnMainPageAndLoggedIn(ctx context.Context) error {
	driver := ctx.Value("driver").(selenium.WebDriver)

	elem, err := driver.FindElement(selenium.ByXPATH, "//*[@id=\"user_links\"]/div/div/h1")
	if err != nil {
		panic(err)
	}

	text, err := elem.Text()
	if err != nil {
		panic(err)
	}

	if !strings.Contains(text, "Главную") {
		panic("we are not on the main page")
	}

	return nil
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.Before(configureDriver)
	ctx.Given(`^I have navigated to main page and not logged in$`, IHaveNavigatedToMainPageAndNotLoggedIn)
	ctx.When(`^I type email (.*)$`, typeEmail)
	ctx.When(`^I type password (.*)$`, typePassword)
	ctx.When(`^I type confirm_password (.*)$`, typeAgainPassword)
	ctx.When("^I clicked Register$", clickRegister)
	ctx.Then(`^I should see the main page and be logged in$`, checkIfOnMainPageAndLoggedIn)
	ctx.After(disableDriver)
}
