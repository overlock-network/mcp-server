package test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestMCPServer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "MCP Server E2E Test Suite")
}
