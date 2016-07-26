package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/dnsimple/dnsimple-go/dnsimple"
)

func main() {
	url := os.Getenv("SNAP_BUILD_URL")
	if len(url) == 0 {
		fmt.Printf("Must provide SNAP_BUILD_URL")
		os.Exit(1)
	}

	oauthToken := os.Getenv("DNSIMPLE_TOKEN")
	if len(oauthToken) == 0 {
		fmt.Printf("Must provide DNSIMPLE_TOKEN")
		os.Exit(1)
	}

	client := dnsimple.NewClient(dnsimple.NewOauthTokenCredentials(oauthToken))

	// get DNSimple client accountID
	whoamiResponse, err := client.Identity.Whoami()
	if err != nil {
		fmt.Printf("Whoami() returned error: %v\n", err)
		os.Exit(1)
	}
	accountID := strconv.Itoa(whoamiResponse.Data.Account.ID)
	zoneID := "snap-telemetry.io"

	result, err := client.Zones.ListRecords(accountID, zoneID, &dnsimple.ZoneRecordListOptions{Name: "latest.snap.ci"})
	if err != nil {
		fmt.Printf("latest.snap.ci.snap-telemetry.io DNS record not found: %v\n", err)
		os.Exit(1)
	}

	record := result.Data[0]

	// replace build.ci.snap-telemtry.io URL record with new s3 folder
	record.Content = url
	client.Zones.UpdateRecord(accountID, zoneID, record.ID, record)
}
