package main

import (
	"flag"
	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	. "github.com/mfamador/go-payments-api/pkg/test"
	log "github.com/sirupsen/logrus"
	"os"
	"testing"
)

var (
	opt        = godog.Options{Output: colors.Colored(os.Stdout)}
	serverURL  *string
	apiVersion *string
)

func init() {
	serverURL = flag.String("server-url", "http://localhost:8080", "the payments server url to test against")
	apiVersion = flag.String("api-version", "v1", "the api version")
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("go-payments-api", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {
	w := NewWorld(*serverURL, *apiVersion)
	s.BeforeScenario(func(interface{}) {
		w.NewData()
		err := DoThen(w.TheServiceIsUp(), func() error {
			return w.ThereAreNoPayments()
		})
		if err != nil {
			log.Info("Error before scenario:")
			log.Info(err)
		}
	})

	s.Step(`^the service is up$`, w.TheServiceIsUp)
	s.Step(`^there are no payments$`, w.ThereAreNoPayments)
	s.Step(`^I query the health endpoint$`, w.IQueryTheHealthEndpoint)
	s.Step(`^I query the metrics endpoint$`, w.IQueryTheMetricsEndpoint)
	s.Step(`^I should have a json$`, w.IShouldHaveAJson)
	s.Step(`^I should have a text$`, w.IShouldHaveAText)
	s.Step(`^I should have status code (\d+)$`, w.IShouldHaveStatusCode)
	s.Step(`^I should have content-type (.*)$`, w.IShouldHaveContentType)
	s.Step(`^that json should have string at (.*) equal to (.*)$`, w.ThatJsonShouldHaveString)
	s.Step(`^that json should have int at (.*) equal to (.*)$`, w.ThatJsonShouldHaveInt)
	s.Step(`^that json should have (\d+) items$`, w.ThatJsonShouldHaveItems)
	s.Step(`^that json should have an (.*)$`, w.ThatJsonShouldHaveA)
	s.Step(`^that json should have a (.*)$`, w.ThatJsonShouldHaveA)
	s.Step(`^that text should match (.*)$`, w.ThatTextShouldMatch)
	s.Step(`^I get all payments$`, w.IGetAllPayments)
	s.Step(`^I get payments (\d+) to (\d+)$`, w.IGetPaymentsFromTo)
	s.Step(`^I get payments without from/to$`, w.IGetPaymentsWithoutFromTo)
	s.Step(`^a payment with id ([a-z]+)$`, w.APaymentWithId)
	s.Step(`^a payment without organisation, and id ([a-z]+)$`, w.APaymentWithIdNoOrganisation)
	s.Step(`^a payment with id ([a-z]+) and amount (.*)$`, w.APaymentWithIdAmount)
	s.Step(`^I create that payment$`, w.ICreateThatPayment)
	s.Step(`^I update that payment$`, w.IUpdateThatPayment)
	s.Step(`^I delete that payment$`, w.IDeleteThatPayment)
	s.Step(`^I get that payment$`, w.IGetThatPayment)
	s.Step(`^I created a new payment with id (.*)$`, w.ICreatedANewPaymentWithId)
	s.Step(`^I created (\d+) payments$`, w.ICreatedPayments)
	s.Step(`^I should have (\d+) payment\(s\)$`, w.IShouldHavePayments)
	s.Step(`^I deleted that payment$`, w.IDeletedThatPayment)
	s.Step(`^I updated that payment$`, w.IUpdatedThatPayment)
	s.Step(`^I delete version (\d+) of that payment$`, w.IDeleteVersionOfThatPayment)
	s.Step(`^I delete that payment, without saying which version$`, w.IDeleteThatPaymentWithoutSayingWhichVersion)
	s.Step(`^I update version (\d+) of that payment$`, w.IUpdateVersionOfThatPayment)
	s.Step(`^that payment has version (\d+)$`, w.ThatPaymentHasVersion)
}
