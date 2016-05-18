package cli

import (
	"fmt"
	"os"
	"path"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func Run() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Usage = "docker management"

	app.Author = ""
	app.Email = ""

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "debug",
			Usage:  "debug mode",
			EnvVar: "DEBUG",
		},
		cli.StringFlag{
			Name:  "log-level, l",
			Value: "info",
			Usage: fmt.Sprintf("Log level (options: debug, info, warn, error, fatal, panic)"),
		},
	}

	// logs
	app.Before = func(c *cli.Context) error {
		log.SetOutput(os.Stderr)
		level, err := log.ParseLevel(c.String("log-level"))
		if err != nil {
			log.Fatalf(err.Error())
		}
		log.SetLevel(level)

		// If a log level wasn't specified and we are running in debug mode,
		// enforce log-level=debug.
		if !c.IsSet("log-level") && !c.IsSet("l") && c.Bool("debug") {
			log.SetLevel(log.DebugLevel)
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:      "server",
			ShortName: "s",
			Usage:     "run daolinet controller",
			Action:    server,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "listen, l",
					Usage: "listen address",
					Value: ":3380",
				},
				cli.StringFlag{
					Name:   "swarm, w",
					Value:  "tcp://127.0.0.1:2375",
					Usage:  "docker swarm addr",
					EnvVar: "DOCKER_HOST",
				},
                cli.StringFlag{
                    Name: "ofc",
                    Value: "http://127.0.0.1:8080",
                    Usage: "openflow controller",
                },
				cli.BoolFlag{
					Name:  "allow-insecure",
					Usage: "enable insecure tls communication",
				},
				flHeartBeat, flDiscoveryOpt,
			},
		},
		{
			Name:      "agent",
			ShortName: "a",
			Usage:     "join manager cluster",
			Action:    agent,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "bridge",
					Usage: "daolinet ovs bridge",
					Value: "daolinet",
				},
				cli.StringFlag{
					Name:  "addr",
					Usage: "discover address",
				},
				cli.StringFlag{
					Name:  "iface",
					Usage: "docker network interface(format <devname:ip>).",
				},
				/*cli.StringFlag{
					Name:  "int-nic",
					Usage: "internal network interface",
				},
				cli.StringFlag{
					Name:  "ext-nic",
					Usage: "public network interface",
				},*/
				flHeartBeat, flTTL, flDiscoveryOpt,
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
