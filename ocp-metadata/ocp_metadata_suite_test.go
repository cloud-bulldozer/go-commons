package ocpmetadata_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestOcpMetadata(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "OcpMetadata Suite")
}
