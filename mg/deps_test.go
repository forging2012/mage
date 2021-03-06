package mg_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/magefile/mage/mg"
	"github.com/pkg/errors"
)

func TestDepsRunOnce(t *testing.T) {
	done := make(chan struct{})
	f := func() {
		done <- struct{}{}
	}
	go mg.Deps(f, f)
	select {
	case <-done:
		// cool
	case <-time.After(time.Millisecond * 100):
		t.Fatal("func not run in a reasonable amount of time.")
	}
	select {
	case <-done:
		t.Fatal("func run twice!")
	case <-time.After(time.Millisecond * 100):
		// cool... this should be plenty of time for the goroutine to have run
	}
}

func TestDepsOfDeps(t *testing.T) {
	ch := make(chan string, 3)
	// this->f->g->h
	h := func() {
		ch <- "h"
	}
	g := func() {
		mg.Deps(h)
		ch <- "g"
	}
	f := func() {
		mg.Deps(g)
		ch <- "f"
	}
	mg.Deps(f)

	res := <-ch + <-ch + <-ch

	if res != "hgf" {
		t.Fatal("expected h then g then f to run, but got " + res)
	}
}

func TestDepError(t *testing.T) {
	// TODO: this test is ugly and relies on implementation details. It should
	// be recreated as a full-stack test.

	f := func() error {
		return errors.New("ouch!")
	}
	defer func() {
		err := recover()
		if err == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(err)
		if "ouch!" != actual {
			t.Fatalf(`expected to get "ouch!" but got "%s"`, actual)
		}
	}()
	mg.Deps(f)
}

func TestDepFatal(t *testing.T) {
	f := func() error {
		return mg.Fatal(99, "ouch!")
	}
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(v)
		if "ouch!" != actual {
			t.Fatalf(`expected to get "ouch!" but got "%s"`, actual)
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", v)
		}
		code := mg.ExitStatus(err)
		if code != 99 {
			t.Fatalf("Expected exit status 99, but got %v", code)
		}
	}()
	mg.Deps(f)
}

func TestDepTwoFatal(t *testing.T) {
	f := func() error {
		return mg.Fatal(99, "ouch!")
	}
	g := func() error {
		return mg.Fatal(11, "bang!")
	}
	defer func() {
		v := recover()
		if v == nil {
			t.Fatal("expected panic, but didn't get one")
		}
		actual := fmt.Sprint(v)
		// order is non-deterministic, so check for both orders
		if "ouch!\nbang!" != actual && "bang!\nouch!" != actual {
			t.Fatalf(`expected to get "ouch!" and "bang!" but got "%s"`, actual)
		}
		err, ok := v.(error)
		if !ok {
			t.Fatalf("expected recovered val to be error but was %T", v)
		}
		code := mg.ExitStatus(err)
		// two different error codes returns, so we give up and just use error
		// code 1.
		if code != 1 {
			t.Fatalf("Expected exit status 1, but got %v", code)
		}
	}()
	mg.Deps(f, g)
}
