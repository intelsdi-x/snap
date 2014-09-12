package collection_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCollection(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Pulse Collection Suite")
}
