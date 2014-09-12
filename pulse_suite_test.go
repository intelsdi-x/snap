package pulse_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPulse(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pulse Suite")
}
