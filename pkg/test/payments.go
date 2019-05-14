package test

import (
	"fmt"
	"github.com/mdaverde/jsonpath"
	. "github.com/smartystreets/assertions"
	"reflect"
)

func (w *World) TheServiceIsUp() error {
	return DoThen(w.IQueryTheHealthEndpoint(), func() error {
		return w.IShouldHaveStatusCode(200)
	})
}

func (w *World) ThereAreNoPayments() error {
	return DoThen(w.IDeleteAllData(), func() error {
		return DoThen(w.IShouldHaveStatusCode(204), func() error {
			return w.IShouldHavePayments(0)
		})
	})
}

func (w *World) IDeleteAllData() error {
	w.Client.Delete("/admin/repo")
	return nil
}

func (w *World) IGetTheRepoInfo() error {
	w.Client.Get("/admin/repo")
	return nil
}

func (w *World) TheRepoShouldHaveItems(expected int) error {
	return DoThen(w.IGetTheRepoInfo(), func() error {
		return DoThen(w.IShouldHaveStatusCode(200), func() error {
			return DoThen(w.IShouldHaveAJson(), func() error {
				return w.ThatJsonShouldHaveInt("count", expected)
			})
		})
	})
}

func (w *World) IQueryTheHealthEndpoint() error {
	w.Client.Get("/health")
	return nil
}

func (w *World) IGetPaymentsWithoutFromTo() error {
	w.Client.Get(w.versionedPath("/payments"))
	return nil
}

func (w *World) IGetAllPayments() error {
	return w.IGetPaymentsFromTo(0, 20) // fetch first 20
}

func (w *World) IGetPaymentsFromTo(from int, to int) error {
	path0 := fmt.Sprintf("/payments?from=%v&to=%v", from, to)
	w.Client.Get(w.versionedPath(path0))
	return nil
}

func (w *World) IQueryTheMetricsEndpoint() error {
	w.Client.Get("/metrics")
	return nil
}

func (w *World) IShouldHaveStatusCode(expected int) error {
	return ExpectThen(ShouldBeNil(w.Client.Err), func() error {
		return ExpectThen(ShouldNotBeNil(w.Client.Resp), func() error {
			return Expect(ShouldEqual(w.Client.Resp.StatusCode, expected))
		})
	})
}

func (w *World) IShouldHaveContentType(expected string) error {
	return ExpectThen(ShouldNotBeNil(w.Client.Resp), func() error {
		return Expect(ShouldContainSubstring(w.Client.Resp.Header.Get("content-type"), expected))
	})
}

func (w *World) IShouldHaveAJson() error {
	return DoThen(w.IShouldHaveContentType("application/json"), func() error {
		return ExpectThen(ShouldNotBeNil(w.Client.Json), func() error {
			w.Data.Subject = w.Client.Json
			return nil
		})
	})
}

func (w *World) IShouldHaveAText() error {
	return ExpectThen(ShouldNotBeNil(w.Client.Text), func() error {
		w.Data.Subject = w.Client.Text
		return nil
	})
}

func (w *World) ThatJsonShouldHaveString(path string, expected string) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		actual, err := jsonpath.Get(w.Data.Subject, path)
		return ExpectThen(ShouldBeNil(err), func() error {
			return Expect(ShouldEqual(actual, expected))
		})
	})
}

func (w *World) ThatJsonShouldHaveInt(path string, expected int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		actual, err := jsonpath.Get(w.Data.Subject, path)
		return ExpectThen(ShouldBeNil(err), func() error {
			return Expect(ShouldEqual(actual, expected))
		})
	})
}

func (w *World) ThatTextShouldMatch(expected string) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		var text string
		return ExpectThen(ShouldEqual(reflect.TypeOf(w.Data.Subject), reflect.TypeOf(text)), func() error {
			text = w.Data.Subject.(string)
			return Expect(ShouldContainSubstring(text, expected))
		})
	})
}

func (w *World) ThatJsonShouldHaveItems(expected int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		actual, err := jsonpath.Get(w.Data.Subject, "data")
		return ExpectThen(ShouldBeNil(err), func() error {
			var items []interface{}
			return ExpectThen(ShouldEqual(reflect.TypeOf(actual), reflect.TypeOf(items)), func() error {
				items =

					actual.([]interface{})
				return Expect(ShouldEqual(len(items), expected))
			})
		})
	})
}

func (w *World) ThatJsonShouldHaveA(path string) error {
	return ExpectThen(ShouldNotBeNil(w.Data.Subject), func() error {
		actual, err := jsonpath.Get(w.Data.Subject, path)
		return ExpectThen(ShouldBeNil(err), func() error {
			return Expect(ShouldNotBeNil(actual))
		})
	})
}

func (w *World) APaymentWithId(id string) error {
	w.Data.PaymentData = &PaymentData{
		Id:           id,
		Version:      0,
		Organisation: "org1",
		Amount:       "1.00",
	}
	return nil
}

func (w *World) APaymentWithIdNoOrganisation(id string) error {
	w.Data.PaymentData = &PaymentData{
		Id:      id,
		Version: 0,
		Amount:  "1.00",
	}
	return nil
}

func (w *World) APaymentWithIdAmount(id string, amount string) error {
	w.Data.PaymentData = &PaymentData{
		Id:           id,
		Version:      0,
		Organisation: "org1",
		Amount:       amount,
	}
	return nil
}

func (w *World) ThatPaymentHasVersion(v int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		w.Data.PaymentData.Version = v
		return nil
	})
}

func (w *World) ICreateThatPayment() error {
	path := w.versionedPath("/payments")
	w.Client.Post(path, w.Data.PaymentData.ToJSON())
	return nil
}

func (w *World) IUpdateThatPayment() error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		path := w.versionedPath(fmt.Sprintf("/payments/%s", p.Id))
		w.Client.Put(path, p.ToJSON())
		return nil
	})
}

func (w *World) IUpdateVersionOfThatPayment(v int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		p.Version = v
		return w.IUpdateThatPayment()
	})
}

func (w *World) IGetThatPayment() error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		path := w.versionedPath(fmt.Sprintf("/payments/%s", p.Id))
		w.Client.Get(path)
		return nil
	})
}

func (w *World) ICreatedANewPaymentWithId(id string) error {
	return DoThen(w.APaymentWithId(id), func() error {
		return DoThen(w.ICreateThatPayment(), func() error {
			return w.IShouldHaveStatusCode(201)
		})
	})
}

func (w *World) ICreatePayments(count int) error {
	return DoSequence(func(it int) error {
		// convert the iteration number into an id
		id := fmt.Sprintf("payment%v", it)
		return DoThen(w.APaymentWithId(id), func() error {
			return DoThen(w.ICreateThatPayment(), func() error {
				return w.IShouldHaveStatusCode(201)
			})
		})
	}, count)
}

func (w *World) ICreatedPayments(count int) error {
	return DoThen(w.ICreatePayments(count), func() error {
		return w.TheRepoShouldHaveItems(count)
	})
}

func (w *World) IDeletedThatPayment() error {
	return DoThen(w.IDeleteThatPayment(), func() error {
		return w.IShouldHaveStatusCode(204)
	})
}

func (w *World) IDeleteThatPayment() error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		path := w.versionedPath(fmt.Sprintf("/payments/%s?version=%v", p.Id, p.Version))
		w.Client.Delete(path)
		return nil
	})
}

func (w *World) IDeleteVersionOfThatPayment(v int) error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		p.Version = v
		return w.IDeleteThatPayment()
	})
}

func (w *World) IDeleteThatPaymentWithoutSayingWhichVersion() error {
	return ExpectThen(ShouldNotBeNil(w.Data.PaymentData), func() error {
		p := w.Data.PaymentData
		path := w.versionedPath(fmt.Sprintf("/payments/%s", p.Id))
		w.Client.Delete(path)
		return nil
	})
}

func (w *World) IUpdatedThatPayment() error {
	return DoThen(w.IUpdateThatPayment(), func() error {
		return w.IShouldHaveStatusCode(200)
	})
}

func (w *World) IShouldHavePayments(expected int) error {
	return DoThen(w.IGetAllPayments(), func() error {
		return DoThen(w.IShouldHaveStatusCode(200), func() error {
			return DoThen(w.IShouldHaveAJson(), func() error {
				return w.ThatJsonShouldHaveItems(expected)
			})
		})
	})
}
