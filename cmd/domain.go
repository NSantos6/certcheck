package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/fatih/color"
	"github.com/NSantos6/certcheck/internal/checker"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var domainWarnDays int
var domainOutputJSON bool
var domainFile string

var domainCmd = &cobra.Command{
	Use:   "domain [domain...]",
	Short: "Check domain registration expiry via WHOIS",
	Example: `  certcheck domain example.com.br
  certcheck domain meusite.com.br outrosite.net.br
  certcheck domain --file domains.yaml --warn-days 60`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domains, err := resolveDomains(args, domainFile)
		if err != nil {
			return err
		}
		if len(domains) == 0 {
			return fmt.Errorf("no domains provided")
		}

		results := runDomainChecks(domains)

		if domainOutputJSON {
			return printDomainJSON(results)
		}
		printDomainTable(results, domainWarnDays)
		return nil
	},
}

func init() {
	domainCmd.Flags().IntVar(&domainWarnDays, "warn-days", 60, "warn when domain expires within N days")
	domainCmd.Flags().BoolVar(&domainOutputJSON, "json", false, "output as JSON")
	domainCmd.Flags().StringVar(&domainFile, "file", "", "file with list of domains (one per line or YAML)")
	rootCmd.AddCommand(domainCmd)
}

func runDomainChecks(domains []string) []checker.DomainResult {
	results := make([]checker.DomainResult, len(domains))
	var wg sync.WaitGroup

	for i, d := range domains {
		wg.Add(1)
		go func(idx int, domain string) {
			defer wg.Done()
			results[idx] = checker.CheckDomain(domain)
		}(i, d)
	}

	wg.Wait()
	return results
}

func printDomainTable(results []checker.DomainResult, warnDays int) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Domain", "Expires", "Days Left", "Registrar", "Status"})
	table.SetBorder(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_RIGHT,
		tablewriter.ALIGN_LEFT,
		tablewriter.ALIGN_LEFT,
	})

	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()

	for _, r := range results {
		if r.Error != nil {
			table.Append([]string{r.Domain, "-", "-", "-", red("ERROR: " + r.Error.Error())})
			continue
		}

		status := green("OK")
		days := fmt.Sprintf("%d", r.DaysLeft)
		if r.DaysLeft < 0 {
			status = red("EXPIRED")
			days = red(days)
		} else if r.DaysLeft <= warnDays {
			status = yellow("EXPIRING SOON")
			days = yellow(days)
		}

		table.Append([]string{
			r.Domain,
			r.ExpiresAt.Format("2006-01-02"),
			days,
			r.Registrar,
			status,
		})
	}

	table.Render()
}

func printDomainJSON(results []checker.DomainResult) error {
	type jsonResult struct {
		Domain    string `json:"domain"`
		ExpiresAt string `json:"expires_at,omitempty"`
		DaysLeft  int    `json:"days_left,omitempty"`
		Registrar string `json:"registrar,omitempty"`
		Error     string `json:"error,omitempty"`
	}

	out := make([]jsonResult, len(results))
	for i, r := range results {
		out[i] = jsonResult{Domain: r.Domain, DaysLeft: r.DaysLeft, Registrar: r.Registrar}
		if r.Error != nil {
			out[i].Error = r.Error.Error()
		} else {
			out[i].ExpiresAt = r.ExpiresAt.Format("2006-01-02")
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
