package main

import (
	"crypto/ed25519"
	"log"
	"time"

	"github.com/fynelabs/selfupdate"
)

func main() {
	done := make(chan struct{}, 2)

	// Used `selfupdatectl create-keys` followed by `selfupdatectl print-key`
	publicKey := ed25519.PublicKey{178, 103, 83, 57, 61, 138, 18, 249, 244, 80, 163, 162, 24, 251, 190, 241, 11, 168, 179, 41, 245, 27, 166, 70, 220, 254, 118, 169, 101, 26, 199, 129}

	// The public key above match the signature of the below file served by our CDN
	httpSource := selfupdate.NewHTTPSource(nil, "http://geoffrey-test-artefacts.fynelabs.com/nomad.exe")
	config := &selfupdate.Config{
		Source: httpSource,
		Schedule: selfupdate.Schedule{
			FetchOnStart: true,
			Interval:     time.Minute * time.Duration(60),
			At: selfupdate.ScheduleAt{
				// Repeated update check hourly at half of the hour
				Repeating: selfupdate.Hourly,
				Time:      time.Date(0, 0, 0, 0, 30, 0, 0, time.UTC)},
		},
		PublicKey: publicKey,

		// This is here to force an update by announcing a time so old that nothing existed
		Current: &selfupdate.Version{Date: time.Unix(100, 0)},

		ProgressCallback: func(f float64, err error) { log.Println("Download", f, "%") },
		RestartConfirmCallback: func() bool {
			done <- struct{}{}
			return true
		},
		UpgradeConfirmCallback: func(_ string) bool { return true },
	}

	_, err := selfupdate.Manage(config)
	if err != nil {
		log.Println("Error while setting up update manager: ", err)
		return
	}

	<-done
}
