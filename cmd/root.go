/*
Copyright © 2024 Michael Bruskov <mixanemca@yandex.ru>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mixanemca/cfdnscli/internal/ui"
	"github.com/spf13/cobra"
	"github.com/version-go/ldflags"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var (
	cfgFile       string
	clientTimeout time.Duration
	content       string
	debug         bool
	name          string
	outputType    string
	proxied       bool
	rrtype        string
	ttl           int
	zone          string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "cfdnscli",
	Short:   "Work with CloudFlare DNS easily from CLI!",
	Version: ldflags.Version(),
	Run:     rootCmdRun,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	build := ldflags.Build()
	vt := rootCmd.VersionTemplate()
	rootCmd.SetVersionTemplate(vt[:len(vt)-1] + " (" + build + ")\n")

	// TODO: add config package
	// cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cfdnscli.yaml)")
	rootCmd.PersistentFlags().DurationVarP(&clientTimeout, "timeout", "T", 5*time.Second, "client timeout")
	rootCmd.PersistentFlags().StringVarP(&outputType, "output-type", "o", "text", "print output in format: text/json")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "turn on debug output to STDERR")

	viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("output-type", rootCmd.PersistentFlags().Lookup("output-type"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	tableStyle := table.DefaultStyles()
	tableStyle.Selected = lipgloss.NewStyle().Background(ui.Color.Highlight)

	// Creates a new table with specified columns and initial empty rows.
	zonesTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 50},
			{Title: "NS", Width: 100},
		}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(50),
		table.WithStyles(tableStyle),
	)
	rrsetTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 50},
			{Title: "TTL", Width: 10},
			{Title: "Type", Width: 10},
			{Title: "Proxied", Width: 10},
			{Title: "Content", Width: 70},
		}),
		table.WithRows([]table.Row{}),
		table.WithFocused(false),
		table.WithHeight(50),
		table.WithStyles(tableStyle),
	)

	m := ui.NewModel()
	m.ClientTimeout = clientTimeout
	m.ZonesTable = zonesTable
	m.RRSetTable = rrsetTable
	m.TableStyle = tableStyle

	// Create a new Bubble Tea program with the model and enable alternate screen
	p := tea.NewProgram(m, tea.WithAltScreen())

	// Run the program and handle any errors
	if _, err := p.Run(); err != nil {
		log.Fatalf("Error running program: %v", err)
	}
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".cfdnscli" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cfdnscli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("ERROR: Config file %s not found", viper.ConfigFileUsed())
		os.Exit(1)
	}
}
