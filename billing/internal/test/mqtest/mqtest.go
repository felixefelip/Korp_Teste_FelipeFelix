// Package mqtest holds the messaging doubles shared by the tests of the web packages.
package mqtest

import (
	"billing/internal/model"
)

type Publisher struct {
	Published []model.Invoice
	Err       error
}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) PublishCloseRequested(invoice model.Invoice) error {
	if p.Err != nil {
		return p.Err
	}

	p.Published = append(p.Published, invoice)

	return nil
}
