package dyndns

import (
	"fmt"
	"net/http"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"

	"context"
	"log"
	"net"
	"os"
	"strings"

	dns "google.golang.org/api/dns/v1"
)

func init() {
	functions.HTTP("handleHTTP", handleHTTP)
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	// Get the IP address of the caller.
	ipAddress := GetCallerIP(r)

	if r.URL.Query().Get("key") != os.Getenv("KEY") {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// Update the Google Cloud DNS A entry for the domain.
	err := UpdateDNSRecord(context.Background(), os.Getenv("PROJECT_ID"), r.URL.Query().Get("zone"), r.URL.Query().Get("domain"), ipAddress)
	if err != nil {
		// Handle the error.
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{ \"status\": \"error\" }"))
		return
	}

	// Write a 200 OK response.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{ \"status\": \"ok\" }"))
}

// GetCallerIP returns the IP address of the caller.
func GetCallerIP(r *http.Request) string {
	// Get the IP address from the X-Forwarded-For header, if present.
	ip := r.Header.Get("X-Forwarded-For")
	if ip != "" {
		// Use the first IP address in the X-Forwarded-For header.
		ips := strings.Split(ip, ",")
		ip = ips[0]
	} else {
		// Get the IP address from the RemoteAddr header.
		ip = r.RemoteAddr
	}

	// Return the IP address.

	_ip := net.ParseIP(ip)
	if _ip == nil {
		splitIp, _, err := net.SplitHostPort(ip)
		if err != nil {
			log.Println(fmt.Errorf("unable to extract ip adress: %w", err))
		}

		_ip = net.ParseIP(splitIp)
	}

	return _ip.String()
}

// UpdateDNSRecord updates the DNS A record for the specified domain with the specified IP address.
func UpdateDNSRecord(ctx context.Context, projectID string, managedZone string, domain string, ipAddress string) error {
	// Create a DNS client.
	client, err := dns.NewService(ctx)
	if err != nil {
		return err
	}

	resourceRecordSetsList, err := client.ResourceRecordSets.List(projectID, managedZone).Do()
	if err != nil {
		return err
	}

	// Prepare a "change" (which is a list of records to add):
	change := &dns.Change{
		Additions: []*dns.ResourceRecordSet{},
	}

	// Set default TTL in case it does not exist
	var ttl int64 = 300
	var exists bool = false

	// See if we already have a DNS entry that matches our record
	for _, resourceRecordSet := range resourceRecordSetsList.Rrsets {
		if resourceRecordSet.Type == "A" && resourceRecordSet.Name == domain {
			if resourceRecordSet.Rrdatas[0] != ipAddress {
				fmt.Printf("found record: %s - %v\n", resourceRecordSet.Name, resourceRecordSet.Rrdatas[0])
				change.Deletions = append(change.Deletions, resourceRecordSet)
				ttl = resourceRecordSet.Ttl
			}

			exists = true
		}
	}

	if (exists && len(change.Deletions) > 0) || !exists {
		change.Additions = append(change.Additions, &dns.ResourceRecordSet{
			Name:    domain,
			Rrdatas: []string{ipAddress},
			Ttl:     ttl,
			Type:    "A",
		})
	}

	if len(change.Additions) > 0 || len(change.Deletions) > 0 {
		changeMade, err := client.Changes.Create(projectID, managedZone, change).Do()
		if err != nil {
			log.Println(fmt.Errorf("unable to make DNS Changes.Create() call! (%w)", err))
			return err
		} else {
			log.Printf("made %v changes to DNS zone (%v), status: %v", len(changeMade.Additions), projectID, changeMade.Status)
		}
	} else {
		log.Printf("no changes to be made")
	}

	return nil
}
