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
	"expl"
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

		host, err := cmd.Flags().GetString("host")
		if err != nil {
			return err
		}

		database, err := cmd.Flags().GetString("database")
		if err != nil {
			return err
		}
		user, err := cmd.Flags().GetString("user")
		if err != nil {
			return err
		}
		pass, err := cmd.Flags().GetString("pass")
		if err != nil {
			return err
		}

		port, err := cmd.Flags().GetInt("port")
		if err != nil {
			return err
		}

		protocol, err := cmd.Flags().GetString("protocol")
		if err != nil {
			return err
		}

		action := args[0]

		switch action {
		case "add":

			err = expl.AddHostAndDatabase(ctx,
				expl.ConfFilePath(filePath),
				expl.DBUser(user),
				expl.DBPass(pass),
				expl.DBHost(host),
				expl.DBDatabase(database),
				expl.DBPort(port),
				expl.DBProtocol(protocol),
			)

			fmt.Printf("conf %s --host %s --database %s --user %s --pass %s --port %d --protocol %s --conf %s\n",
				action, host, database, user, pass, port, protocol, filePath,
			)
		case "rm":

			err = expl.RemoveHostAndDatabase(ctx,
				expl.ConfFilePath(filePath),
				expl.DBUser(user),
				expl.DBPass(pass),
				expl.DBHost(host),
				expl.DBDatabase(database),
				expl.DBPort(port),
				expl.DBProtocol(protocol),
			)

			fmt.Printf("conf %s --host %s --database %s --user %s --pass %s --port %d --protocol %s --conf %s\n",
				action, host, database, user, pass, port, protocol, filePath,
			)
		case "mapping":
			err = expl.ReloadAllTableInfo(ctx, filePath)
			fmt.Printf("conf %s\n", action)
		default:
			return fmt.Errorf("%s is invalid action\n", action)
		}

		if err != nil {
			fmt.Print(expl.Message(err))
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(confCmd)

	template := `expl conf action [parameter]...
  action:
	add:     add database and host, user, password... in setting used to explain sql.
	rm:      remove database and host, user, password... in setting used to explain sql.
	mapping: create or update table-dtabase mapping. using sing Host and Databse settings created by the above add action command.

  option:
	-d --database: string  database. used by any sqls. (using onlh simple mode. not using in table mapping mode)
	-H --host:     string  host used by any sqls.(using onlh simple mode. not using table in mapping mode)
	-u --user:     string  database user used by any sqls.(using onlh simple mode. not using table in mapping mode)
	-p --pass:     string  database password used by any sqls.(using onlh simple mode. not using table in mapping mode)
	-c, --conf:	   string  config file. it includes table mapping. You can set and use "EXPL_CONF" environment variable as a default value.
                     "EXPL_CONF" environment variable as a default value.
                     value:
                       [mapping file path]: database-table mapping file path. default file is ./table_map.yaml.
                     ex)
                       -c $GOPATH/bin/table-mapping.yaml

	ex)
	  expl conf add --host localhost --database database1 --user root --pass password -conf configpath
	  expl conf rm --host localhost --database database2 --user root --pass password -conf configpath
	  expl conf mapping -conf configpath  // make table-database mapping file in database1 and database2.
`

	confCmd.SetUsageTemplate(template)
	confCmd.SetHelpTemplate(template)

	confCmd.Flags().StringP("database", "d", "localhost", "database")
	confCmd.Flags().StringP("host", "H", "", "host")
	confCmd.Flags().StringP("user", "u", "", "database user")
	confCmd.Flags().StringP("pass", "p", "", "database password")
	confCmd.Flags().IntP("port", "P", 3306, "database port")
	confCmd.Flags().StringP("protocol", "R", "tcp", "database protocol. (default:tcp)")

	confCmd.Flags().StringP("conf", "c", "", "config. which includes database-table mapping file.")
}
