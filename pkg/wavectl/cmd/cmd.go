package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/wujie1993/waves/pkg/wavectl"
)

func Execute() {
	getCmd := NewResourceCmd()
	getCmd.Use = "get [RESOURCE TYPE] [RESOURCE NAME]"
	getCmd.Short = "Get resources"
	getCmd.Args = cobra.MinimumNArgs(1)
	getCmd.Run = func(cmd *cobra.Command, args []string) {
		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		format, err := cmd.Flags().GetString("format")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		level, err := cmd.Flags().GetInt("level")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		log.SetLevel(log.Level(level))

		var resource string
		var resourceName string
		if len(args) > 0 {
			resource = args[0]
		}
		if len(args) > 1 {
			resourceName = args[1]
		}

		wavectl.GetResource(wavectl.GetResourceOptions{
			Endpoint:     endpoint,
			Namespace:    namespace,
			Resource:     resource,
			ResourceName: resourceName,
			Format:       format,
		})
	}

	createCmd := NewResourceCmd()
	createCmd.Use = "create"
	createCmd.Short = "Create resources"
	createCmd.MarkFlagRequired("file")
	createCmd.Run = func(cmd *cobra.Command, args []string) {
		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		level, err := cmd.Flags().GetInt("level")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		log.SetLevel(log.Level(level))

		wavectl.CreateResource(wavectl.CreateResourceOptions{
			Endpoint: endpoint,
			File:     file,
		})
	}

	applyCmd := NewResourceCmd()
	applyCmd.Use = "apply"
	applyCmd.Short = "Update resources"
	applyCmd.Run = func(cmd *cobra.Command, args []string) {
		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		level, err := cmd.Flags().GetInt("level")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		log.SetLevel(log.Level(level))

		wavectl.ApplyResource(wavectl.ApplyResourceOptions{
			Endpoint: endpoint,
			File:     file,
		})
	}

	deleteCmd := NewResourceCmd()
	deleteCmd.Use = "delete"
	deleteCmd.Short = "Delete resources"
	deleteCmd.Run = func(cmd *cobra.Command, args []string) {
		endpoint, err := cmd.Flags().GetString("endpoint")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		namespace, err := cmd.Flags().GetString("namespace")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		var resource string
		var resourceName string
		if len(args) > 0 {
			resource = args[0]
		}
		if len(args) > 1 {
			resourceName = args[1]
		}

		file, err := cmd.Flags().GetString("file")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		level, err := cmd.Flags().GetInt("level")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		log.SetLevel(log.Level(level))

		wavectl.DeleteResource(wavectl.DeleteResourceOptions{
			Endpoint:     endpoint,
			Namespace:    namespace,
			Resource:     resource,
			ResourceName: resourceName,
			File:         file,
		})
	}

	hostPluginCmd := &cobra.Command{
		Use:   "hostplugin [install|uninstall]",
		Short: "Manage host plugins",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			endpoint, err := cmd.Flags().GetString("endpoint")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			host, err := cmd.Flags().GetString("host")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			pluginName, err := cmd.Flags().GetString("plugin-name")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			pluginVersion, err := cmd.Flags().GetString("plugin-version")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			force, err := cmd.Flags().GetBool("force")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			level, err := cmd.Flags().GetInt("level")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			log.SetLevel(log.Level(level))

			wavectl.ManageHostPlugin(wavectl.HostPluginOptions{
				Endpoint:      endpoint,
				Action:        args[0],
				Host:          host,
				PluginName:    pluginName,
				PluginVersion: pluginVersion,
				Force:         force,
			})
		},
	}
	hostPluginCmd.Flags().StringP("endpoint", "e", "http://127.0.0.1:8000/deployer", "api endpoint of visible deploy platform")
	hostPluginCmd.Flags().IntP("level", "l", 0, "logs level(0.Panic|1.Fatal|2.Error|3.Warn|4.Info|5.Debug|6.Trace)")
	hostPluginCmd.Flags().StringP("host", "", "", "the destnation host to manage plugins")
	hostPluginCmd.Flags().StringP("plugin-name", "", "", "the plugin name")
	hostPluginCmd.Flags().StringP("plugin-version", "", "", "the plugin version")
	hostPluginCmd.Flags().BoolP("force", "", false, "ignore operation restrictions")
	hostPluginCmd.MarkFlagRequired("host")
	hostPluginCmd.MarkFlagRequired("plugin-name")

	rootCmd := &cobra.Command{
		Use:   "wavectl",
		Short: "The command line tool of visible deploy platform",
	}
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(hostPluginCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewResourceCmd() *cobra.Command {
	cmd := new(cobra.Command)
	cmd.Flags().StringP("endpoint", "e", "http://127.0.0.1:8000/deployer", "api endpoint of visible deploy platform")
	cmd.Flags().StringP("namespace", "n", "default", "the namespace to which the resource belongs")
	cmd.Flags().StringP("file", "f", "", "the local path of the loaded resource")
	cmd.Flags().StringP("format", "", "table", "resource output format")
	cmd.Flags().IntP("level", "l", 0, "logs level(0.Panic|1.Fatal|2.Error|3.Warn|4.Info|5.Debug|6.Trace)")
	return cmd
}
