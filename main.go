package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"
)

// Config holds values to substitute into templates
type Config struct {
	Name    string // project name
	Port    string // HTTP server port
	DLVPort string // Delve debugger port
}

//go:embed templates/**
//go:embed templates/**/*
var tmplFS embed.FS

var (
	// cobra root command
	rootCmd = &cobra.Command{
		Use:   "generator",
		Short: "Scaffold a Go project from embedded templates",
		Long: `generator uses embedded files in the templates directory
	to create a new project, substituting name, port, and Delve port.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGenerator()
		},
	}

	// flags
	flagName    string
	flagPort    string
	flagDLVPort string
	flagOut     string
)

func init() {
	// register flags on rootCmd
	rootCmd.Flags().StringVar(&flagName, "name", "demo", "Project name (and module path)")
	rootCmd.Flags().StringVar(&flagPort, "port", "8080", "Server HTTP port")
	rootCmd.Flags().StringVar(&flagDLVPort, "dlv-port", "2345", "Delve debug port")
	rootCmd.Flags().StringVar(&flagOut, "out", ".", "Destination directory for generated project")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runGenerator() error {
	cfg := Config{flagName, flagPort, flagDLVPort}

	// create project directory
	projDir := filepath.Join(flagOut, cfg.Name)
	if err := os.MkdirAll(projDir, 0755); err != nil {
		return fmt.Errorf("error creating project dir: %w", err)
	}

	// Walk embedded templates recursively
	err := fs.WalkDir(tmplFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rawRel := strings.TrimPrefix(path, "templates/")
		if rawRel == "" {
			return nil // skip templates root
		}

		// Determine output path, stripping .tmpl suffix if present
		outRel := rawRel
		if strings.HasSuffix(rawRel, ".tmpl") {
			outRel = strings.TrimSuffix(rawRel, ".tmpl")
		}
		outPath := filepath.Join(projDir, outRel)

		if d.IsDir() {
			return os.MkdirAll(outPath, 0755)
		}

		// Read embedded file
		raw, err := tmplFS.ReadFile(path)
		if err != nil {
			return err
		}

		// Parse and execute template
		tmpl, err := template.New(rawRel).Parse(string(raw))
		if err != nil {
			return fmt.Errorf("template parse %s: %w", rawRel, err)
		}

		f, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer f.Close()

		if err := tmpl.Execute(f, cfg); err != nil {
			return fmt.Errorf("template exec %s: %w", rawRel, err)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error generating project: %w", err)
	}

	fmt.Printf("Project '%s' generated in %s\n", cfg.Name, projDir)
	return nil
}
