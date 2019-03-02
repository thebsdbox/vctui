package cmd

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/thebsdbox/vctui/pkg/vctui"
	"github.com/vmware/govmomi"

	"github.com/spf13/cobra"
)

type vc struct {
	address  string
	username string
	password string
	insecure bool
}

var vCenterDetails vc

// Release - this struct contains the release information populated when building katbox
var Release struct {
	Version string
	Build   string
}
var logLevel int

func init() {

	// VMware vCenter details
	vctuiCmd.Flags().StringVar(&vCenterDetails.address, "address", os.Getenv("VCURL"), "The Url/address of a VMware vCenter server")
	vctuiCmd.Flags().StringVar(&vCenterDetails.username, "user", os.Getenv("VCUSER"), "The Url/address of a VMware vCenter server")
	vctuiCmd.Flags().StringVar(&vCenterDetails.password, "pass", os.Getenv("VCPASS"), "The Url/address of a VMware vCenter server")
	vctuiCmd.Flags().BoolVar(&vCenterDetails.insecure, "insecure", false, "The Url/address of a VMware vCenter server")

	vctuiCmd.PersistentFlags().IntVar(&logLevel, "logLevel", 5, "Set the logging level [0=panic, 3=warning, 5=debug]")
	vctuiCmd.AddCommand(vctuiVersion)
	log.SetLevel(log.Level(logLevel))
}

func parseCredentials(v *vc) (*url.URL, error) {

	// Check that an address was actually entered
	if v.address == "" {
		return nil, fmt.Errorf("No VMware vCenter URL/Address has been submitted")
	}

	// Check that the URL can be parsed
	u, err := url.Parse(v.address)
	if err != nil {
		return nil, fmt.Errorf("URL can't be parsed, ensure it is https://username:password/<address>/sdk")
	}

	// Check if a username was entered
	if v.username == "" {
		// if no username does one exist as part of the url
		if u.User.Username() == "" {
			return nil, fmt.Errorf("No VMware vCenter Username has been submitted")
		}
	} else {
		// A username was submitted update the url
		u.User = url.User(v.username)
	}

	if v.password == "" {
		_, set := u.User.Password()
		if set == false {
			return nil, fmt.Errorf("No VMware vCenter Password has been submitted")
		}
	} else {
		u.User = url.UserPassword(u.User.Username(), v.password)
	}
	return u, nil
}

var vctuiCmd = &cobra.Command{
	Use:   "vctui",
	Short: "VMware vCenter Text User Interface",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetLevel(log.Level(logLevel))

		u, err := parseCredentials(&vCenterDetails)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c, err := govmomi.NewClient(ctx, u, true)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		defer c.Logout(ctx)

		vms, err := vctui.VMInventory(c)
		if err != nil {
			log.Fatalf("%s", err.Error())
		}

		// Sort function to sort by name
		sort.Slice(vms, func(i, j int) bool {
			switch strings.Compare(vms[i].Name(), vms[j].Name()) {
			case -1:
				return true
			case 1:
				return false
			}
			return vms[i].Name() > vms[j].Name()
		})

		vctui.MainUI(vms)
		return
	},
}

// Execute - starts the command parsing process
func Execute() {
	if os.Getenv("VCLOG") != "" {
		i, err := strconv.ParseInt(os.Getenv("VCLOG"), 10, 8)
		if err != nil {
			log.Fatalf("Error parsing environment variable [VCLOG")
		}
		// We've only parsed to an 8bit integer, however i is still a int64 so needs casting
		logLevel = int(i)
	} else {
		// Default to logging anything Info and below
		logLevel = int(log.InfoLevel)
	}

	if err := vctuiCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var vctuiVersion = &cobra.Command{
	Use:   "version",
	Short: "Version and Release information about the plunder tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Plunder Release Information\n")
		fmt.Printf("Version:  %s\n", Release.Version)
		fmt.Printf("Build:    %s\n", Release.Build)
	},
}
