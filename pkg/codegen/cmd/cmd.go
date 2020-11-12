package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/wujie1993/waves/pkg/codegen"
)

var rootCmd = &cobra.Command{
	Use:   "codegen",
	Short: "The code generator for visible deploy platform",
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Generate code files of api layer",
}

var ormCmd = &cobra.Command{
	Use:   "orm",
	Short: "Generate code files of orm layer",
	Run: func(cmd *cobra.Command, args []string) {
		pkgPath, err := cmd.Flags().GetString("pkg-path")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		codegen.GenOrm(codegen.GenOrmOptions{
			PkgPath: pkgPath,
		})
	},
}

func init() {
	ormCmd.Flags().StringP("pkg-path", "p", "", "package path to generate")
	ormCmd.MarkFlagRequired("pkg-path")

	rootCmd.AddCommand(ormCmd)
	rootCmd.AddCommand(apiCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
