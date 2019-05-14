package payments

import (
	"encoding/json"
	"fmt"
	. "github.com/mfamador/go-payments-api/pkg/util"
	"github.com/pkg/errors"
	"strconv"
	"strings"
)

type PaymentAttributes struct {
	Amount string `json:"amount"`
}

func (pa *PaymentAttributes) Validate() error {

	amount, err := strconv.ParseFloat(pa.Amount, 64)
	if err != nil {
		return errors.Wrap(err, "Invalid payment amount")
	}

	if amount <= 0 {
		return fmt.Errorf("Payment amount must be positive")
	}

	return nil
}

type Payment struct {
	Id           string            `json:"id"`
	Type         string            `json:"type"`
	Version      int               `json:"version"`
	Organisation string            `json:"organisation_id"`
	Attributes   PaymentAttributes `json:"attributes"`
}

func (p *Payment) Validate() error {

	if len(strings.TrimSpace(p.Id)) == 0 {
		return errors.New("Id is empty")
	}

	if p.Type != "Payment" {
		return fmt.Errorf("Invalid type: %s", p.Type)
	}

	if len(strings.TrimSpace(p.Organisation)) == 0 {
		return errors.New("Organisation is empty")
	}

	return p.Attributes.Validate()
}

func (p *Payment) ToRepoItem() (*RepoItem, error) {
	repoItem := &RepoItem{
		Id:           p.Id,
		Version:      p.Version,
		Organisation: p.Organisation,
	}
	bytes, err := json.Marshal(p.Attributes)
	if err != nil {
		return repoItem, errors.Wrap(err, "Unable to serialize payment attributes")
	}

	repoItem.Attributes = string(bytes)
	return repoItem, nil
}

func NewPaymentFromRepoItem(item *RepoItem) (*Payment, error) {
	p := &Payment{
		Type:         "Payment",
		Id:           item.Id,
		Version:      item.Version,
		Organisation: item.Organisation,
	}

	var attrs PaymentAttributes
	if item.Attributes != "" {
		err := json.NewDecoder(strings.NewReader(item.Attributes)).Decode(&attrs)
		if err != nil {
			return p, errors.Wrap(err, "Error parsing repo item attributes")
		}
	}
	p.Attributes = attrs

	return p, nil
}

func NewPaymentsFromRepoItems(items []*RepoItem) ([]*Payment, error) {
	payments := []*Payment{}
	for _, i := range items {
		p, err := NewPaymentFromRepoItem(i)
		if err != nil {
			return payments, err
		}
		payments = append(payments, p)
	}

	return payments, nil
}

type PaymentRequest struct {
	Payment *Payment `json:"data"`
}

type PaymentResponse struct {
	Data  *Payment `json:"data"`
	Links Links    `json:"links"`
}

type Links map[string]string

type PaymentsResponse struct {
	Data  []*Payment `json:"data"`
	Links Links      `json:"links"`
}
