package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lanrat/czds"
)

// flags
var (
	username    = flag.String("username", "", "username to authenticate with")
	password    = flag.String("password", "", "password to authenticate with")
	verbose     = flag.Bool("verbose", false, "enable verbose logging")
	id          = flag.String("id", "", "ID of specific zone request to lookup, defaults to printing all")
	zone        = flag.String("zone", "", "same as -id, but prints the request by zone name")
	showVersion = flag.Bool("version", false, "print version and exit")
)

var (
	version = "unknown"
	client  *czds.Client
)

func v(format string, v ...interface{}) {
	if *verbose {
		log.Printf(format, v...)
	}
}

func checkFlags() {
	flag.Parse()
	if *showVersion {
		fmt.Printf("Version: %s\n", version)
		os.Exit(0)
	}
	flagError := false
	if len(*username) == 0 {
		log.Printf("must pass username")
		flagError = true
	}
	if len(*password) == 0 {
		log.Printf("must pass password")
		flagError = true
	}
	if flagError {
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	checkFlags()

	client = czds.NewClient(*username, *password)
	if *verbose {
		client.SetLogger(log.Default())
	}

	// validate credentials
	v("Authenticating to %s", client.AuthURL)
	err := client.Authenticate()
	if err != nil {
		log.Fatal(err)
	}

	if *zone != "" {
		// get id from zone name
		zoneID, err := client.GetZoneRequestID(*zone)
		if err != nil {
			log.Fatal(err)
		}
		id = &zoneID
	}

	if *id == "" {
		listAll()
		return
	}

	info, err := client.GetRequestInfo(*id)
	if err != nil {
		log.Fatal(err)
	}
	printRequestInfo(info)
}

func printRequestInfo(info *czds.RequestsInfo) {
	fmt.Printf("ID:\t%s\n", info.RequestID)
	fmt.Printf("TLD:\t%s (%s)\n", info.TLD.TLD, info.TLD.ULabel)
	fmt.Printf("Status:\t%s\n", info.Status)
	fmt.Printf("Created:\t%s\n", info.Created.Format(time.ANSIC))
	fmt.Printf("Updated:\t%s\n", info.LastUpdated.Format(time.ANSIC))
	fmt.Printf("Expires:\t%s\n", expiredTime(info.Expired))
	fmt.Printf("AutoRenew:\t%t\n", info.AutoRenew)
	fmt.Printf("Extensible:\t%t\n", info.Extensible)
	fmt.Printf("ExtensionInProcess:\t%t\n", info.ExtensionInProcess)
	fmt.Printf("Cancellable:\t%t\n", info.Cancellable)
	fmt.Printf("Request IP:\t%s\n", info.RequestIP)
	fmt.Println("FTP IPs:\t", info.FtpIps)
	fmt.Printf("Reason:\t%s\n", info.Reason)
	fmt.Printf("History:\n")
	for _, event := range info.History {
		fmt.Printf("\t%s\t%s\n", event.Timestamp.Format(time.ANSIC), event.Action)
	}
}

func listAll() {
	requests, err := client.GetAllRequests(czds.RequestAll)
	if err != nil {
		log.Fatal(err)
	}

	v("Total requests: %d", len(requests))
	if len(requests) > 0 {
		printHeader()
		for _, request := range requests {
			printRequest(request)
		}
	}
}

func printRequest(request czds.Request) {
	fmt.Printf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%t\n",
		request.TLD,
		request.RequestID,
		request.ULabel,
		request.Status,
		request.Created.Format(time.ANSIC),
		request.LastUpdated.Format(time.ANSIC),
		expiredTime(request.Expired),
		request.SFTP)
}

func printHeader() {
	fmt.Printf("TLD\tID\tUnicodeTLD\tStatus\tCreated\tUpdated\tExpires\tSFTP\n")
}

func expiredTime(t time.Time) string {
	if t.Unix() > 0 {
		return t.Format(time.ANSIC)
	}
	return ""
}
