package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/izhan05803/DATABASE_CURATED_IN_GO/internal/database"
)

func main() {
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)

	listCategory := listCmd.String("category", "", "filter by category: 'API Reference', 'Developer Guide', 'Tutorial', 'Specification'")

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	db := database.New()

	switch os.Args[1] {
	case "list":
		listCmd.Parse(os.Args[2:])
		var entries []database.Entry
		if *listCategory != "" {
			entries = db.FilterByCategory(database.Category(*listCategory))
		} else {
			entries = db.All()
		}
		printEntries(entries)

	case "search":
		searchCmd.Parse(os.Args[2:])
		query := strings.Join(searchCmd.Args(), " ")
		if query == "" {
			fmt.Fprintln(os.Stderr, "error: search requires a query argument")
			os.Exit(1)
		}
		results := db.Search(query)
		if len(results) == 0 {
			fmt.Println("No entries found matching:", query)
			return
		}
		printEntries(results)

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printEntries(entries []database.Entry) {
	if len(entries) == 0 {
		fmt.Println("No entries found.")
		return
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tCATEGORY\tURL")
	fmt.Fprintln(w, "--\t-----\t--------\t---")
	for _, e := range entries {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", e.ID, e.Title, e.Category, e.URL)
	}
	w.Flush()
	fmt.Printf("\n%d resource(s) found.\n", len(entries))
}

func printUsage() {
	fmt.Println(`DATABASE CURATED IN GO — curated technical documentation database

Usage:
  go run . <command> [flags]

Commands:
  list               List all curated entries
    -category string   Filter by category (optional)
                       Values: "API Reference", "Developer Guide", "Tutorial", "Specification"

  search <query>     Search entries by title, description, or tag

Examples:
  go run . list
  go run . list -category "API Reference"
  go run . search go modules
  go run . search rest api`)
}
