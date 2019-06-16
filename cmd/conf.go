// Copyright © 2019 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"errors"
	"expl/service"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// confCmd represents the conf command
var confCmd = &cobra.Command{
	Use:   "conf",
	Short: "conf action",
	Long: `conf action [option]
			action:
				add: add database hosting
				rm: rm database hosting
		`,

	//SilenceUsage: true,

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("requires action")
		}
		return nil
	},

	RunE: func(cmd *cobra.Command, args []string) error {

		ctx := context.Background()

		// config file path
		filePath, err := cmd.Flags().GetString("conf")
		if err != nil {
			return err
		}
		if filePath == "" {
			filePath = os.Getenv("EXPL_CONF")
		}

		action := args[0]

		switch action {
		case "add":
			host := args[1]
			database := args[2]
			user := args[3]
			pass := args[4]

			err = service.AddHostAndDatabase(ctx, user, pass, host, database, filePath)
			fmt.Printf("conf %s %s %s %s %s %s\n", action, host, database, user, pass, filePath)
		case "rm":
			host := args[1]
			database := args[2]
			user := args[3]
			pass := args[4]

			err = service.RemoveHostAndDatabase(ctx, user, pass, host, database, filePath)
			fmt.Printf("conf %s %s %s %s %s %s\n", action, host, database, user, pass, filePath)
		case "mapping":
			err = service.ReloadAllTableInfo(ctx, filePath)
			fmt.Printf("conf %s\n", action)
		default:
			return fmt.Errorf("%s is invalid action\n", action)
		}

		if err != nil {
			fmt.Print(service.Message(err))
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(confCmd)

	template := `expl conf action [parameter]...
  action:
	add:     add database and host, user, password in setting used to explain sql.
	         ex) add host databasea user password 
	rm:      remove database and host, user, password in setting used to explain sql.
	         ex) rm host database user password
	mapping: create or update table-dtabase mapping. using sing Host and Databse settings created by the above add action command.

  parameter:
	ex)
	  expl conf add localhost database1 root password -c configpath
	  expl conf rm localhost database2 root password -c configpath
	  expl conf mapping -c configpath  // make table-database mapping file in database1 and database2.

  option:
	-c, --conf:	config file. You can set and use "EXPL_CONF" environment variable as a default value.
`

	confCmd.SetUsageTemplate(template)
	confCmd.SetHelpTemplate(template)

	confCmd.Flags().StringP("conf", "c", "", "config. which includes database-table mapping file.")
}
