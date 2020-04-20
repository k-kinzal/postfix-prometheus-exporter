package util_test

import (
	"fmt"
	"github.com/k-kinzal/postfix-prometheus-exporter/util"
	"testing"
)

func ExampleEmailMask() {
	masked := util.EmailMask("foo@example.com")
	fmt.Println(masked)
	// Output: ***@example.com
}

func TestEmailMask(t *testing.T) {
	masked := util.EmailMask("foo1234@example.com")
	if masked != "***@example.com" {
		t.Errorf("expected `***@example.com`, but actual is `%s`", masked)
	}
}

func TestEmailMaskUsePeriod(t *testing.T) {
	masked := util.EmailMask("foo.1234@example.com")
	if masked != "***@example.com" {
		t.Errorf("expected `***@example.com`, but actual is `%s`", masked)
	}
}

func TestEmailMaskUsePlus(t *testing.T) {
	masked := util.EmailMask("foo+1234@example.com")
	if masked != "***@example.com" {
		t.Errorf("expected `***@example.com`, but actual is `%s`", masked)
	}
}

func TestEmailMaskUseSubDomain(t *testing.T) {
	masked := util.EmailMask("foo@sub.example.com")
	if masked != "***@sub.example.com" {
		t.Errorf("expected `***@sub.example.com`, but actual is `%s`", masked)
	}
}
