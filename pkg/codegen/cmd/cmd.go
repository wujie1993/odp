package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/wujie1993/waves/pkg/codegen"
)

func Execute() {
	// 初始化api子命令
	apiCmd := &cobra.Command{
		Use:   "api",
		Short: "Generate code files of api layer",
	}

	// 初始化orm子命令
	ormCmd := &cobra.Command{
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
	ormCmd.Flags().StringP("pkg-path", "p", "", "package path to generate")
	ormCmd.MarkFlagRequired("pkg-path")

	// 初始化client子命令
	cliCmd := &cobra.Command{
		Use:   "client",
		Short: "Generate code files of client set",
		Run: func(cmd *cobra.Command, args []string) {
			input, err := cmd.Flags().GetString("input")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			output, err := cmd.Flags().GetString("output")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			codegen.GenClient(codegen.GenClientOptions{
				InputPkgPath:  input,
				OutputPkgPath: output,
			})
		},
	}
	cliCmd.Flags().StringP("input", "i", "", "The path to scan objects")
	cliCmd.MarkFlagRequired("input")
	cliCmd.Flags().StringP("output", "o", "", "The path to generate code files")
	cliCmd.MarkFlagRequired("output")

	// 初始化codegen命令
	rootCmd := &cobra.Command{
		Use:   "codegen",
		Short: "The code generator for visible deploy platform",
	}
	rootCmd.AddCommand(ormCmd)
	rootCmd.AddCommand(apiCmd)
	rootCmd.AddCommand(cliCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
