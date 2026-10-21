package google

import (
	"testing"

	"github.com/orangedeng/autok3s/pkg/types"
	"github.com/orangedeng/autok3s/pkg/utils"

	"github.com/spf13/cobra"
)

func TestFlagDefinitions(t *testing.T) {
	p := newProvider()

	flagSets := map[string][]types.Flag{
		"create":     p.GetCreateFlags(),
		"join":       p.GetJoinFlags(),
		"option":     p.GetOptionFlags(),
		"delete":     p.GetDeleteFlags(),
		"ssh":        p.GetSSHFlags(),
		"credential": p.GetCredentialFlags(),
	}

	for set, flags := range flagSets {
		t.Run(set, func(t *testing.T) {
			seen := map[string]bool{}
			for _, f := range flags {
				if f.Name == "" {
					t.Fatalf("flag with empty name in %s flags", set)
				}
				if f.P == nil {
					t.Fatalf("flag %s has no value pointer", f.Name)
				}
				if seen[f.Name] {
					t.Fatalf("duplicate flag %s in %s flags", f.Name, set)
				}
				seen[f.Name] = true
			}

			// every declared flag must be bindable by the CLI, i.e. its value
			// type is one of the types ConvertFlags knows how to register.
			cmd := &cobra.Command{Use: "test"}
			utils.ConvertFlags(cmd, flags)
			for name := range seen {
				if cmd.Flags().Lookup(name) == nil {
					t.Fatalf("flag %s is not registered by ConvertFlags", name)
				}
			}
		})
	}
}

func TestUsageExamples(t *testing.T) {
	p := newProvider()
	for _, action := range []string{"create", "join", "delete", "ssh"} {
		if p.GetUsageExample(action) == "" {
			t.Fatalf("missing usage example for action %s", action)
		}
	}
}
