// Copyright (c) 2025 Grant Carthew
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/p3bot/snag/internal/browser"
	"github.com/p3bot/snag/internal/format"
	"github.com/p3bot/snag/internal/logger"
	"github.com/p3bot/snag/internal/validate"
)

var Version = "dev"

const (
	ExitCodeSuccess   = 0
	ExitCodeError     = 1
	ExitCodeInterrupt = 130 // 128 + SIGINT (2)
	ExitCodeSIGTERM   = 143 // 128 + SIGTERM (15)
)

const (
	MaxDisplayURLLength = 80
	MaxTabLineLength    = 120
)

const (
	DefaultTimeout = 30
)

type Config struct {
	URL           string
	OutputFile    string
	OutputDir     string
	Format        string
	Timeout       int
	WaitFor       string
	Port          int
	CloseTab      bool
	ForceHeadless bool
	OpenBrowser   bool
	UserAgent     string
	UserDataDir   string
	TempProfile   bool
}

func (c *Config) BrowserOptions() browser.BrowserOptions {
	return browser.BrowserOptions{
		Port:          c.Port,
		ForceHeadless: c.ForceHeadless,
		OpenBrowser:   c.OpenBrowser,
		UserAgent:     c.UserAgent,
		UserDataDir:   c.UserDataDir,
		TempProfile:   c.TempProfile,
	}
}

// browserOptionsFromFlags validates profile and user-agent flags, then builds
// launch options through Config.BrowserOptions so every start path shares one mapping.
func browserOptionsFromFlags(cmd *cobra.Command, openBrowser, forceHeadless bool) (browser.BrowserOptions, error) {
	validatedUserAgent := validate.UserAgent(userAgent, cmd.Flags().Changed("user-agent"))
	if !trustedSession {
		return browser.BrowserOptions{
			ForceHeadless: true,
			UserAgent:     validatedUserAgent,
			TempProfile:   true,
		}, nil
	}

	validatedUserDataDir := ""
	if !tempProfile && cmd.Flags().Changed("user-data-dir") {
		validatedDir, err := validate.UserDataDir(userDataDir)
		if err != nil {
			return browser.BrowserOptions{}, err
		}
		validatedUserDataDir = validatedDir
	}

	cfg := &Config{
		Port:          port,
		ForceHeadless: forceHeadless,
		OpenBrowser:   openBrowser,
		UserAgent:     validatedUserAgent,
		UserDataDir:   validatedUserDataDir,
		TempProfile:   tempProfile,
	}
	return cfg.BrowserOptions(), nil
}

var (
	urlFile        string
	flagOutput     string
	outputDir      string
	flagFormat     string
	timeout        int
	waitFor        string
	port           int
	closeTab       bool
	forceHead      bool
	openBrowser    bool
	listTabs       bool
	tab            string
	allTabs        bool
	killBrowser    bool
	runDoctor      bool
	showVersion    bool
	info           bool
	verbose        bool
	debug          bool
	colorMode      string
	userAgent      string
	userDataDir    string
	tempProfile    bool
	trustedSession bool
	skillPrint     bool
)

var helpTemplate = fmt.Sprintf(`USAGE:
  snag [options] URL

DESCRIPTION:
  Markdown-first Chromium renderer for public HTTP(S) pages. Normal mode uses
  a sandboxed headless browser, ephemeral profile, public-network policy, and a
  random loopback CDP port. Browser profiles and existing tabs require the
  explicit --trusted-session mode.

EXAMPLES:
  snag https://example.com
  snag --format html https://example.com
  snag --format text --timeout 30 https://example.com
  snag --wait-for ".content" https://example.com
  snag -d pages/ https://example.com https://go.dev
  snag --format png -o page.png https://example.com

OPTIONS:
  -f, --format string          Output: md | html | text | pdf | png (default md)
  -i, --info                   Output page metadata as JSON
  -o, --output string          Save output to a file
  -d, --output-dir string      Save auto-named files to a directory
      --url-file string        Read URLs from a file or stdin with "-"
      --timeout int            Page load timeout in seconds (default %d)
  -w, --wait-for string        Wait for a CSS selector
      --user-agent string      Override the browser user agent
      --trusted-session        Permit explicit profile and existing-tab access
  -b, --open-browser           Open a trusted visible browser
  -l, --list-tabs              List tabs in a trusted browser
  -t, --tab string             Fetch from a trusted browser tab
  -a, --all-tabs               Fetch all trusted browser tabs
      --user-data-dir string   Trusted Chromium profile directory
  -p, --port int               Trusted browser CDP port (default 9222)
      --skill                  Print the default-reader agent skill
      --color string           stderr colour: auto | always | never
      --debug                  Enable debug diagnostics
      --verbose                Enable verbose diagnostics
  -h, --help                   Help
  -v, --version                Version

Session options fail closed unless --trusted-session is present. Local files,
private-network destinations, process killing, update checks, and remote skill
catalogs remain disabled.
`, DefaultTimeout)

var rootCmd = &cobra.Command{
	Use:           "snag [options] URL...",
	Short:         "Intelligently fetch web page content using a browser engine",
	Args:          cobra.ArbitraryArgs,
	RunE:          runCobra,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	rootCmd.Flags().StringVar(&urlFile, "url-file", "", "Read URLs from file (one per line, supports comments)")
	rootCmd.Flags().StringVarP(&flagOutput, "output", "o", "", "Save output to file instead of stdout")
	rootCmd.Flags().StringVarP(&outputDir, "output-dir", "d", "", "Save files with auto-generated names to directory")
	rootCmd.Flags().StringVarP(&flagFormat, "format", "f", format.Markdown, "Output format: md | html | text | pdf | png")
	rootCmd.Flags().StringVarP(&waitFor, "wait-for", "w", "", "Wait for CSS selector before extracting content")
	rootCmd.Flags().StringVarP(&tab, "tab", "t", "", "Fetch from existing tab by pattern (tab number or string)")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "", "Override the browser user agent")
	rootCmd.Flags().StringVar(&userDataDir, "user-data-dir", "", "Chromium user data directory (overrides the persistent snag profile)")
	rootCmd.Flags().BoolVar(&tempProfile, "temp-profile", false, "Ephemeral profile for this launch (mutually exclusive with --user-data-dir)")

	rootCmd.Flags().IntVar(&timeout, "timeout", DefaultTimeout, "Page load timeout in seconds")
	rootCmd.Flags().IntVarP(&port, "port", "p", 9222, "Trusted Chromium/Chrome remote debugging port")

	rootCmd.Flags().BoolVarP(&closeTab, "close-tab", "c", false, "Close the browser tab after fetching content")
	rootCmd.Flags().BoolVar(&forceHead, "force-headless", false, "Force headless mode even if the browser is running")
	rootCmd.Flags().BoolVarP(&openBrowser, "open-browser", "b", false, "Open browser visibly with remote debugging enabled (no URL required)")
	rootCmd.Flags().BoolVarP(&listTabs, "list-tabs", "l", false, "List all open tabs in the browser")
	rootCmd.Flags().BoolVarP(&allTabs, "all-tabs", "a", false, "Process all open browser tabs (saves with auto-generated filenames)")
	rootCmd.Flags().BoolVarP(&killBrowser, "kill-browser", "k", false, "Kill browser processes with remote debugging enabled")
	rootCmd.Flags().BoolVar(&runDoctor, "doctor", false, "Display comprehensive diagnostic information")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVarP(&info, "info", "i", false, "Output page metadata as JSON (title, URL, domain, slug, timestamp)")
	rootCmd.Flags().BoolVar(&verbose, "verbose", false, "Enable verbose logging output (connection, fetch, and batch progress)")
	rootCmd.Flags().BoolVar(&debug, "debug", false, "Enable debug output")
	rootCmd.Flags().StringVar(&colorMode, "color", logger.ColorAuto, "Colour stderr diagnostics: auto | always | never")
	rootCmd.Flags().BoolVar(&trustedSession, "trusted-session", false, "Permit explicit profile and existing-tab access")

	rootCmd.Flags().BoolVar(&skillPrint, "skill", false, "Print the embedded agent skill contract")

	rootCmd.MarkFlagsMutuallyExclusive("verbose", "debug")
	rootCmd.MarkFlagsMutuallyExclusive("user-data-dir", "temp-profile")

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		useColor, err := logger.ResolveColor(colorMode)
		if err != nil {
			return err
		}
		logger.SetDefault(logger.NewWithWriter(logLevel(), os.Stderr, useColor))
		return nil
	}

	rootCmd.SetHelpTemplate(helpTemplate)
}

func validateFlagCombinations(cmd *cobra.Command, hasURLs bool, hasMultipleURLs bool) error {
	if cmd.Flags().Changed("tab") && hasURLs {
		logger.Error("Cannot use both --tab and URL arguments (mutually exclusive content sources)")
		return ErrTabURLConflict
	}

	if allTabs && hasURLs {
		logger.Error("Cannot use both --all-tabs and URL arguments (mutually exclusive content sources)")
		return fmt.Errorf("conflicting flags: --all-tabs and URL arguments")
	}

	if cmd.Flags().Changed("tab") && allTabs {
		logger.Error("Cannot use both --tab and --all-tabs (mutually exclusive content sources)")
		return fmt.Errorf("conflicting flags: --tab and --all-tabs")
	}

	if openBrowser && forceHead {
		logger.Error("Cannot use both --force-headless and --open-browser (conflicting modes)")
		return fmt.Errorf("conflicting flags: --force-headless and --open-browser")
	}

	if forceHead && cmd.Flags().Changed("tab") {
		logger.Error("Cannot use --force-headless with --tab (--tab requires existing browser connection)")
		return fmt.Errorf("conflicting flags: --force-headless and --tab")
	}

	if forceHead && allTabs {
		logger.Error("Cannot use --force-headless with --all-tabs (--all-tabs requires existing browser connection)")
		return fmt.Errorf("conflicting flags: --force-headless and --all-tabs")
	}

	outputFile := strings.TrimSpace(flagOutput)
	outDir := strings.TrimSpace(outputDir)

	if outputFile != "" && outDir != "" {
		logger.Error("Cannot use both --output and --output-dir")
		logger.Info("Use --output for specific filename OR --output-dir for auto-generated filename")
		return fmt.Errorf("conflicting flags: --output and --output-dir")
	}

	if hasMultipleURLs && outputFile != "" {
		logger.Error("Cannot use --output with multiple content sources. Use --output-dir instead")
		return ErrOutputFlagConflict
	}

	if allTabs && outputFile != "" {
		logger.Error("Cannot use --output with multiple content sources. Use --output-dir instead")
		return ErrOutputFlagConflict
	}

	if closeTab && forceHead {
		logger.Warning("--close-tab is ignored in headless mode (tabs close automatically)")
	}

	if openBrowser && cmd.Flags().Changed("tab") {
		logger.Warning("--tab ignored with --open-browser (no content fetching)")
	}

	if openBrowser && allTabs {
		logger.Warning("--all-tabs ignored with --open-browser (no content fetching)")
	}

	if info && cmd.Flags().Changed("format") {
		logger.Error("Cannot use both --info and --format (--info always outputs JSON)")
		return fmt.Errorf("conflicting flags: --info and --format")
	}

	if info && outDir != "" {
		logger.Error("Cannot use --output-dir with --info (use --output for single file)")
		return fmt.Errorf("conflicting flags: --info and --output-dir")
	}

	if info && hasMultipleURLs {
		logger.Error("Cannot use --info with multiple URLs (single URL only)")
		return fmt.Errorf("conflicting flags: --info and multiple URLs")
	}

	if info && allTabs {
		logger.Error("Cannot use --info with --all-tabs (single content source only)")
		return fmt.Errorf("conflicting flags: --info and --all-tabs")
	}

	return nil
}

func logLevel() logger.LogLevel {
	if debug {
		return logger.LevelDebug
	}
	if verbose {
		return logger.LevelVerbose
	}
	return logger.LevelNormal
}

func runCobra(cmd *cobra.Command, args []string) error {
	skillActive := skillFlagsChanged(cmd)

	if showVersion {
		fmt.Printf("snag version %s\n", Version)
		fmt.Println("Repository: https://github.com/jacobbello-pg/snaggle")
		return nil
	}

	if skillActive {
		return handleSkill(cmd, args)
	}

	if err := enforceTrustBoundary(cmd); err != nil {
		return err
	}

	var urls []string

	// Load URLs from file if specified
	if urlFile != "" {
		fileURLs, err := loadURLsFromFile(cmd.Context(), strings.TrimSpace(urlFile))
		if err != nil {
			return err
		}
		urls = append(urls, fileURLs...)
	}

	for _, arg := range args {
		trimmedArg := strings.TrimSpace(arg)
		if trimmedArg != "" {
			urls = append(urls, trimmedArg)
		}
	}

	if runDoctor {
		return handleDoctor(cmd)
	}

	if killBrowser {
		if len(urls) > 0 {
			logger.Error("Cannot use --kill-browser with URL arguments (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and URL arguments")
		}
		if listTabs {
			logger.Error("Cannot use --kill-browser with --list-tabs (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --list-tabs")
		}
		if allTabs {
			logger.Error("Cannot use --kill-browser with --all-tabs (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --all-tabs")
		}
		if cmd.Flags().Changed("tab") {
			logger.Error("Cannot use --kill-browser with --tab (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --tab")
		}
		if openBrowser {
			logger.Error("Cannot use --kill-browser with --open-browser (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --open-browser")
		}
		if urlFile != "" {
			logger.Error("Cannot use --kill-browser with --url-file (conflicting operations)")
			return fmt.Errorf("conflicting flags: --kill-browser and --url-file")
		}
		return handleKillBrowser(cmd)
	}

	if listTabs {
		if len(urls) > 0 {
			logger.Verbose("--list-tabs overrides URL arguments (URLs will be ignored)")
		}
		return handleListTabs(cmd)
	}

	hasURLs := len(urls) > 0
	hasMultipleURLs := len(urls) > 1
	if err := validateFlagCombinations(cmd, hasURLs, hasMultipleURLs); err != nil {
		return err
	}

	if info {
		if cmd.Flags().Changed("tab") {
			return handleInfoFromTab(cmd)
		}
		if len(urls) == 1 {
			return handleInfoFromURL(cmd, urls[0])
		}
		logger.Error("--info requires exactly one URL or --tab")
		return fmt.Errorf("--info requires exactly one URL or --tab")
	}

	if allTabs {
		return handleAllTabs(cmd)
	}

	if cmd.Flags().Changed("tab") {
		return handleTabFetch(cmd)
	}

	if openBrowser && len(urls) == 0 {
		if cmd.Flags().Changed("format") {
			logger.Warning("--format ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("output") {
			logger.Warning("--output ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("output-dir") {
			logger.Warning("--output-dir ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("timeout") {
			logger.Warning("--timeout ignored with --open-browser (no content fetching)")
		}
		if cmd.Flags().Changed("wait-for") {
			logger.Warning("--wait-for ignored with --open-browser (no content fetching)")
		}
		if closeTab {
			logger.Warning("--close-tab ignored with --open-browser (no content fetching)")
		}

		opts, err := browserOptionsFromFlags(cmd, true, false)
		if err != nil {
			return err
		}

		logger.Verbose("Opening browser...")
		bm := browser.NewBrowserManager(opts)
		return bm.OpenBrowserOnly(cmd.Context())
	}

	if len(urls) == 0 {
		logger.Error("No URLs provided")
		logger.ErrorWithSuggestion("Provide URLs as arguments or use --url-file", "snag <url> or snag --url-file urls.txt")
		return validate.ErrNoValidURLs
	}

	if openBrowser && len(urls) > 0 {
		return handleOpenURLsInBrowser(cmd, urls)
	}

	if len(urls) == 1 {
		urlStr := urls[0]

		validatedURL, err := validate.URL(urlStr)
		if err != nil {
			return err
		}

		logger.Verbose("Target URL: %s", validatedURL)

		cfg, err := newEmitConfig(cmd, false)
		if err != nil {
			return err
		}

		opts, err := browserOptionsFromFlags(cmd, openBrowser, forceHead)
		if err != nil {
			return err
		}

		cfg.URL = validatedURL
		cfg.Port = opts.Port
		cfg.ForceHeadless = opts.ForceHeadless
		cfg.OpenBrowser = opts.OpenBrowser
		cfg.UserAgent = opts.UserAgent
		cfg.UserDataDir = opts.UserDataDir
		cfg.TempProfile = opts.TempProfile

		logger.Debug("Config: format=%s, timeout=%d, port=%d", cfg.Format, cfg.Timeout, cfg.Port)
		logger.Verbose("Configuration: format=%s, timeout=%ds, port=%d", cfg.Format, cfg.Timeout, cfg.Port)

		return snag(cmd.Context(), cfg)
	}

	return handleMultipleURLs(cmd, urls)
}

func enforceTrustBoundary(cmd *cobra.Command) error {
	if killBrowser {
		return fmt.Errorf("--kill-browser is disabled in Snaggle")
	}
	if trustedSession {
		return nil
	}
	if runDoctor {
		return nil
	}

	for _, name := range []string{"all-tabs", "list-tabs", "open-browser", "port", "tab", "user-data-dir"} {
		if cmd.Flags().Changed(name) {
			return fmt.Errorf("--%s requires --trusted-session", name)
		}
	}
	return nil
}
