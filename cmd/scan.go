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

var scanWarnDaysSSL int
var scanWarnDaysDomain int
var scanOutputJSON bool
var scanFile string

var scanCmd = &cobra.Command{
	Use:   "scan [domain...]",
	Short: "Check both SSL and domain registration expiry",
	Example: `  certcheck scan example.com.br
  certcheck scan --file domains.yaml
  certcheck scan example.com --ssl-warn-days 14 --domain-warn-days 90`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domains, err := resolveDomains(args, scanFile)
		if err != nil {
			return err
		}
		if len(domains) == 0 {
			return fmt.Errorf("no domains provided")
		}

		type combined struct {
			ssl    checker.SSLResult
			domain checker.DomainResult
		}

		results := make([]combined, len(domains))
		var wg sync.WaitGroup

		for i, d := range domains {
			wg.Add(1)
			go func(idx int, domain string) {
				defer wg.Done()
				var inner sync.WaitGroup
				inner.Add(2)
				go func() {
					defer inner.Done()
					results[idx].ssl = checker.CheckSSL(domain)
				}()
				go func() {
					defer inner.Done()
					results[idx].domain = checker.CheckDomain(domain)
				}()
				inner.Wait()
			}(i, d)
		}

		wg.Wait()

		if scanOutputJSON {
			type jsonRow struct {
				Domain string `json:"domain"`
				SSL    struct {
					ExpiresAt string `json:"expires_at,omitempty"`
					DaysLeft  int    `json:"days_left,omitempty"`
					Issuer    string `json:"issuer,omitempty"`
					Error     string `json:"error,omitempty"`
				} `json:"ssl"`
				Registration struct {
					ExpiresAt string `json:"expires_at,omitempty"`
					DaysLeft  int    `json:"days_left,omitempty"`
					Registrar string `json:"registrar,omitempty"`
					Error     string `json:"error,omitempty"`
				} `json:"registration"`
			}

			out := make([]jsonRow, len(results))
			for i, r := range results {
				out[i].Domain = domains[i]
				if r.ssl.Error != nil {
					out[i].SSL.Error = r.ssl.Error.Error()
				} else {
					out[i].SSL.ExpiresAt = r.ssl.ExpiresAt.Format("2006-01-02")
					out[i].SSL.DaysLeft = r.ssl.DaysLeft
					out[i].SSL.Issuer = r.ssl.Issuer
				}
				if r.domain.Error != nil {
					out[i].Registration.Error = r.domain.Error.Error()
				} else {
					out[i].Registration.ExpiresAt = r.domain.ExpiresAt.Format("2006-01-02")
					out[i].Registration.DaysLeft = r.domain.DaysLeft
					out[i].Registration.Registrar = r.domain.Registrar
				}
			}

			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(out)
		}

		red := color.New(color.FgRed).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		green := color.New(color.FgGreen).SprintFunc()

		statusFor := func(daysLeft int, warnDays int, err error) (string, string) {
			if err != nil {
				return "-", red("ERROR")
			}
			days := fmt.Sprintf("%d", daysLeft)
			if daysLeft < 0 {
				return red(days), red("EXPIRED")
			} else if daysLeft <= warnDays {
				return yellow(days), yellow("EXPIRING SOON")
			}
			return days, green("OK")
		}

		fmt.Println()
		fmt.Println(color.New(color.Bold).Sprint("SSL Certificates"))
		sslTable := tablewriter.NewWriter(os.Stdout)
		sslTable.SetHeader([]string{"Domain", "Expires", "Days Left", "Issuer", "Status"})
		sslTable.SetBorder(false)
		sslTable.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		for _, r := range results {
			days, status := statusFor(r.ssl.DaysLeft, scanWarnDaysSSL, r.ssl.Error)
			issuer, expiry := r.ssl.Issuer, "-"
			if r.ssl.Error == nil {
				expiry = r.ssl.ExpiresAt.Format("2006-01-02")
			} else {
				issuer = r.ssl.Error.Error()
			}
			sslTable.Append([]string{r.ssl.Domain, expiry, days, issuer, status})
		}
		sslTable.Render()

		fmt.Println()
		fmt.Println(color.New(color.Bold).Sprint("Domain Registration"))
		domTable := tablewriter.NewWriter(os.Stdout)
		domTable.SetHeader([]string{"Domain", "Expires", "Days Left", "Registrar", "Status"})
		domTable.SetBorder(false)
		domTable.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		for _, r := range results {
			days, status := statusFor(r.domain.DaysLeft, scanWarnDaysDomain, r.domain.Error)
			registrar, expiry := r.domain.Registrar, "-"
			if r.domain.Error == nil {
				expiry = r.domain.ExpiresAt.Format("2006-01-02")
			} else {
				registrar = r.domain.Error.Error()
			}
			domTable.Append([]string{r.domain.Domain, expiry, days, registrar, status})
		}
		domTable.Render()
		fmt.Println()

		return nil
	},
}

func init() {
	scanCmd.Flags().IntVar(&scanWarnDaysSSL, "ssl-warn-days", 30, "warn when SSL cert expires within N days")
	scanCmd.Flags().IntVar(&scanWarnDaysDomain, "domain-warn-days", 60, "warn when domain registration expires within N days")
	scanCmd.Flags().BoolVar(&scanOutputJSON, "json", false, "output as JSON")
	scanCmd.Flags().StringVar(&scanFile, "file", "", "file with list of domains (one per line or YAML)")
	rootCmd.AddCommand(scanCmd)
}
