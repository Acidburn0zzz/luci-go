// Copyright 2016 The LUCI Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"fmt"

	"github.com/maruel/subcommands"
	"golang.org/x/net/context"

	"github.com/luci/luci-go/client/flagpb"
	"github.com/luci/luci-go/common/auth"
	"github.com/luci/luci-go/common/cli"
	"github.com/luci/luci-go/grpc/prpc"
)

const (
	cmdF2JUsage = `f2j [flags] <server> <message type> <message flags>

  server: host ("example.com") or port for localhost (":8080").
  message type: full name of the message type.
  message flags: a message in flagpb format.
`

	cmdF2JDesc = "converts a message from flagpb format to JSON format."
)

func cmdF2J(defaultAuthOpts auth.Options) *subcommands.Command {
	return &subcommands.Command{
		UsageLine: cmdF2JUsage,
		ShortDesc: cmdF2JDesc,
		LongDesc: `Converts a message from flagpb format to JSON format.
It is convenient for switching from flag format in "rpc call" to json format
once a command line becomes unreadable.

Example:

  $ rpc fmt f2j :8080 helloworld.HelloRequest -name Lucy
  {
    "name: "Lucy"
  }

See also j2f subcommand.
`,
		CommandRun: func() subcommands.CommandRun {
			c := &f2jRun{}
			c.registerBaseFlags(defaultAuthOpts)
			return c
		},
	}
}

type f2jRun struct {
	cmdRun
}

func (r *f2jRun) Run(a subcommands.Application, args []string, env subcommands.Env) int {
	if len(args) < 2 {
		return r.argErr(cmdF2JDesc, cmdF2JUsage, "")
	}
	host, msgType := args[0], args[1]
	args = args[2:]

	ctx := cli.GetContext(a, r, env)
	client, err := r.authenticatedClient(ctx, host)
	if err != nil {
		return ecAuthenticatedClientError
	}
	return r.done(flagsToJSON(ctx, client, msgType, args))
}

// flagsToJSON loads server description, resolves msgType, parses
// args to a message according to msgType and prints the message in JSON
// format.
func flagsToJSON(c context.Context, client *prpc.Client, msgType string, args []string) error {
	// Load server description.
	serverDesc, err := loadDescription(c, client)
	if err != nil {
		return err
	}

	// Resolve message type.
	desc, err := serverDesc.resolveMessage(msgType)
	if err != nil {
		return err
	}

	// Parse flags.
	msg, err := flagpb.UnmarshalUntyped(args, desc, flagpb.NewResolver(serverDesc.Description))
	if err != nil {
		return err
	}

	// Print JSON.
	jsonBytes, err := json.MarshalIndent(msg, "", " ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", jsonBytes)
	return nil
}
