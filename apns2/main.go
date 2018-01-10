package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/token"
)

var (
	certificatePath = flag.String("certificate-path", "", "Path to certificate file.")
	authTokenPath   = flag.String("auth-token-path", "", "Path to auth token file.")
	authTokenKeyID  = flag.String("auth-token-key", "", "Auth token Key ID (required with auth-token-path)")
	authTokenTeamID = flag.String("auth-token-team", "", "Auth token Team ID (required with auth-token-path)")
	topic           = flag.String("topic", "", "The topic of the remote notification, which is typically the bundle ID for your app")
	mode            = flag.String("mode", "production", "APNS server to send notifications to. `production` or `development`.")
)

func usage() {
	fmt.Fprint(os.Stderr, `Listens to STDIN to send notifications and writes APNS response code and reason to STDOUT.
Certificate or AuthToken are required. The expected format is: <DeviceToken> <APNS Payload>
Example: aff0c63d9eaa63ad161bafee732d5bc2c31f66d552054718ff19ce314371e5d0 {"aps": {"alert": "hi"}}

`)
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	var client *apns2.Client

	if *authTokenPath != "" {
		if *authTokenKeyID == "" || *authTokenTeamID == "" {
			log.Fatalf("Must specify token-key and token-team when using token based auth")
		}
		authKey, err := token.AuthKeyFromFile(*authTokenPath)
		if err != nil {
			log.Fatalf("Error retrieving auth-token `%v`: %v", *authTokenPath, err)
		}
		token := &token.Token{
			AuthKey: authKey,
			KeyID:   *authTokenKeyID,
			TeamID:  *authTokenTeamID,
		}
		client = apns2.NewTokenClient(token)
	}
	if *certificatePath != "" {
		cert, err := certificate.FromPemFile(*certificatePath, "")
		if err != nil {
			log.Fatalf("Error retrieving certificate `%v`: %v", *certificatePath, err)
		}
		client = apns2.NewClient(cert)
	}
	if client == nil {
		usage()
	}

	if *mode == "development" {
		client.Development()
	} else {
		client.Production()
	}

	scanner := bufio.NewScanner(os.Stdin)

	for scanner.Scan() {
		in := scanner.Text()
		notificationArgs := strings.SplitN(in, " ", 2)
		token := notificationArgs[0]
		payload := notificationArgs[1]

		notification := &apns2.Notification{
			DeviceToken: token,
			Topic:       *topic,
			Payload:     payload,
		}

		res, err := client.Push(notification)

		if err != nil {
			log.Fatal("Error: ", err)
		} else {
			fmt.Printf("%v: '%v'\n", res.StatusCode, res.Reason)
		}
	}
}
