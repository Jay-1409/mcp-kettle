package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"mcp-kettel/internal/generate"
	"mcp-kettel/internal/scan"
	"mcp-kettel/internal/scan/express"
	"mcp-kettel/internal/scan/fastapi"
	"mcp-kettel/internal/tui"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "mcp-kettel:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	flags := flag.NewFlagSet("mcp-kettel", flag.ContinueOnError)
	input := flags.String("input", "", "project directory to scan for FastAPI or Express routes")
	output := flags.String("output", "", "generated MCP server directory (default: <host>/mcp-server)")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *input == "" || flags.NArg() != 0 {
		return fmt.Errorf("usage: mcp-kettel --input DIR [--output DIR]")
	}
	host, err := filepath.Abs(*input)
	if err != nil {
		return err
	}
	info, err := os.Stat(host)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("host path is not a directory: %s", host)
	}

	candidates, err := scan.Directory(host, fastapi.Scan, scan.Files(express.ScanFile))
	if err != nil {
		return err
	}
	if len(candidates) == 0 {
		return fmt.Errorf("no supported FastAPI or Express routes found")
	}
	selected, cancelled, err := tui.Select(candidates)
	if err != nil {
		return err
	}
	if cancelled {
		fmt.Println("Cancelled; no files written.")
		return nil
	}
	if *output == "" {
		*output = filepath.Join(host, "mcp-server")
	} else if *output, err = filepath.Abs(*output); err != nil {
		return err
	}
	if err := generate.Write(*output, selected); err != nil {
		return err
	}
	fmt.Printf("Generated %d MCP tools in %s\n", len(selected), *output)
	return nil
}
