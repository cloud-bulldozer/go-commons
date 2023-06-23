package indexers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestIndexers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Indexers Suite")

}
