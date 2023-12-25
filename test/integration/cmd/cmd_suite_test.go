package cmd_test

import (
	"testing"

	"github.com/cnrancher/autok3s/cmd"
	"github.com/cnrancher/autok3s/cmd/functional"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var rootCmd *cobra.Command

func TestCmd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Suite")
}

var _ = BeforeSuite(func() {
	rootCmd = cmd.Command()
	rootCmd.AddCommand(
		functional.Commands...,
	)

	Expect(rootCmd).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	rootCmd = nil
})
