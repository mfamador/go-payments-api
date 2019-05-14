package test

import (
	"errors"
	"fmt"
)

func ExpectThen(msg string, next func() error) error {
	if msg != "" {
		return errors.New(msg)
	} else {
		return next()
	}
}

func Expect(msg string) error {
	if msg != "" {
		return errors.New(msg)
	} else {
		return nil
	}
}

func DoThen(err error, next func() error) error {
	if err != nil {
		return err
	} else {
		return next()
	}
}

func DoSequence(step func(it int) error, count int) error {
	for i := 0; i < count; i++ {
		err := step(i)
		if err != nil {
			return err
		}
	}
	return nil
}

type PaymentData struct {
	Id           string
	Version      int
	Organisation string
	Amount       string
}

func (p *PaymentData) ToJSON() string {
	return fmt.Sprintf(`{ 
		"data": {
			"id": "%s",
			"type": "Payment",
			"version": %v,
			"organisation_id": "%s",
			"attributes": {
				"amount": "%s"
			}
		}
	}`, p.Id, p.Version, p.Organisation, p.Amount)
}

type ScenarioData struct {
	PaymentData *PaymentData
	Subject     interface{}
}

type World struct {
	serverUrl  string
	apiVersion string
	Client     *Client
	Data       *ScenarioData
}

func NewWorld(serverUrl string, apiVersion string) *World {
	return &World{
		serverUrl:  serverUrl,
		apiVersion: apiVersion,
	}
}

func (w *World) NewData() {
	w.Data = &ScenarioData{}
	w.Client = NewClient(w.serverUrl)
}

func (w *World) versionedPath(path string) string {
	return fmt.Sprintf("/%s%s", w.apiVersion, path)
}
