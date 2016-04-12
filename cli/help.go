package cli

import (
    "os"
    "path"

    "github.com/codegangsta/cli"
)

func init() {
    // See https://github.com/codegangsta/cli/pull/171/files
    cli.CommandHelpTemplate = `{{$DISCOVERY := or (eq .Name "server") (eq .Name "agent")}}Usage: ` + path.Base(os.Args[0]) + `{{.Name}}{{if .Flags}} [OPTIONS]{{end}} {{if $DISCOVERY}}<discovery>{{end}}

{{.Usage}}{{if $DISCOVERY}}

Arguments:
   <discovery>    discovery service to use [$DAOLI_DISCOVERY]
                   * token://<token>
                   * consul://<ip>/<path>
                   * etcd://<ip1>,<ip2>/<path>
                   * file://path/to/file
                   * zk://<ip1>,<ip2>/<path>
                   * [nodes://]<ip1>,<ip2>{{end}}{{if .Flags}}

Options:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`
}
