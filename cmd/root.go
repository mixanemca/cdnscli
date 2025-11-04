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

// Package cmd holds cdnscli command and it sub-commands.
package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mixanemca/cdnscli/internal/config"
	pp "github.com/mixanemca/cdnscli/internal/prettyprint"
	"github.com/mixanemca/cdnscli/internal/ui"
	"github.com/mixanemca/cdnscli/internal/ui/theme"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thediveo/enumflag/v2"
	"github.com/version-go/ldflags"
)

var (
	cfgFile       string
	clientTimeout time.Duration
	content       string
	debug         bool
	name          string
	proxied       bool
	rrtype        string
	ttl           int
	zone          string
	appConfig     *config.Config
)

// define output format with default
var outputFormat pp.OutputFormat = pp.FormatText

var outputFormatList = map[pp.OutputFormat][]string{
	pp.FormatText: {"text"},
	pp.FormatJSON: {"json"},
	pp.FormatNone: {"none"},
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "cdnscli",
	Short:   "Cloud DNS CLI - manage DNS records across multiple providers",
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

	// Initialize config on command execution
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cdnscli.yaml)")
	rootCmd.PersistentFlags().DurationVarP(&clientTimeout, "timeout", "T", 10*time.Second, "client timeout")
	rootCmd.PersistentFlags().VarP(
		enumflag.New(&outputFormat, "output-format", outputFormatList, enumflag.EnumCaseSensitive),
		"output-format", "o", "print output in format: text/json/none",
	)
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "turn on debug output to STDERR")

	if err := viper.BindPFlag("timeout", rootCmd.PersistentFlags().Lookup("timeout")); err != nil {
		log.Fatalf("Failed bind flag %q: %v", "timeout", err)
	}
	if err := viper.BindPFlag("output-format", rootCmd.PersistentFlags().Lookup("output-format")); err != nil {
		log.Fatalf("Failed bind flag %q: %v", "output-format", err)
	}
	if err := viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug")); err != nil {
		log.Fatalf("Failed bind flag %q: %v", "debug", err)
	}
	if err := viper.BindPFlag("client_timeout", rootCmd.PersistentFlags().Lookup("timeout")); err != nil {
		log.Fatalf("Failed bind flag %q: %v", "client_timeout", err)
	}
}

// getTimeout returns the timeout to use, checking config first, then flag, then default.
func getTimeout() time.Duration {
	if appConfig != nil && clientTimeout == 0 {
		return appConfig.GetClientTimeout()
	}
	if clientTimeout != 0 {
		return clientTimeout
	}
	return 10 * time.Second
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Failed to load config: %v\n", err)
		// Continue with default config
		cfg = &config.Config{
			ClientTimeout: 10 * time.Second,
			OutputFormat:  "text",
			Debug:          false,
			Providers:     make(map[string]config.ProviderConfig),
		}
	}

	// Override with command line flags if set
	if clientTimeout != 0 {
		cfg.ClientTimeout = clientTimeout
		viper.Set("client_timeout", clientTimeout)
	}
	if outputFormat != pp.FormatText {
		// Convert OutputFormat to string using the format list
		if formats, ok := outputFormatList[outputFormat]; ok && len(formats) > 0 {
			cfg.OutputFormat = formats[0]
			viper.Set("output_format", formats[0])
		}
	}
	if debug {
		cfg.Debug = debug
		viper.Set("debug", debug)
	}

	appConfig = cfg

	// Validate config
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: Config validation failed: %v\n", err)
		// Continue anyway - validation errors might be non-critical
	}
}

func rootCmdRun(cmd *cobra.Command, args []string) {
	tableStyle := table.DefaultStyles()
	tableStyle.Selected = lipgloss.NewStyle().Background(theme.Color.Highlight)

	// Creates a new table with specified columns and initial empty rows.
	zonesTable := table.New(
		table.WithColumns([]table.Column{
			{Title: "Name", Width: 50},
			{Title: "NS", Width: 100},
			{Title: "Provider", Width: 20},
		}),
		table.WithRows([]table.Row{}),
		table.WithFocused(true),
		table.WithHeight(35),
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
		table.WithHeight(35),
		table.WithStyles(tableStyle),
	)

	m := ui.NewModel()
	m.ClientTimeout = getTimeout()
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

