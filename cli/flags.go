package cli

import (
	"github.com/codegangsta/cli"
	"os"
)

func getDiscovery(c *cli.Context) string {
	if len(c.Args()) == 1 {
		return c.Args()[0]
	}
	return os.Getenv("DAOLI_DISCOVERY")
}

var (
	flHeartBeat = cli.StringFlag{
		Name:  "heartbeat",
		Value: "60s",
		Usage: "period between each heartbeat",
	}
	flTTL = cli.StringFlag{
		Name:  "ttl",
		Value: "180s",
		Usage: "sets the expiration of an ephemeral node",
	}
	flTimeout = cli.StringFlag{
		Name:  "timeout",
		Value: "10s",
		Usage: "timeout period",
	}
	flTLS = cli.BoolFlag{
		Name:  "tls",
		Usage: "use TLS; implied by --tlsverify=true",
	}
	flTLSCaCert = cli.StringFlag{
		Name:  "tlscacert",
		Usage: "trust only remotes providing a certificate signed by the CA given here",
	}
	flTLSCert = cli.StringFlag{
		Name:  "tlscert",
		Usage: "path to TLS certificate file",
	}
	flTLSKey = cli.StringFlag{
		Name:  "tlskey",
		Usage: "path to TLS key file",
	}
	flTLSVerify = cli.BoolFlag{
		Name:  "tlsverify",
		Usage: "use TLS and verify the remote",
	}
	flDiscoveryOpt = cli.StringSliceFlag{
		Name:  "discovery-opt",
		Usage: "discovery options",
		Value: &cli.StringSlice{},
	}
)
