package cmd

import (
	"fmt"
	"os"

	"github.com/rifelpet/iam-policy-builder/pkg/parser"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aws-policy-builder [projectpath]",
	Short: "aws-policy-builder creates IAM policies based on go source code",
	Long: `A source code parser that creates IAM policies based on the
								aws-go-sdk usage.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectPath := args[0]
		err := parser.Parse(projectPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	},
}

// Execute parses out cobra flags and runs the policy builder
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
