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
	"expl/model"
	"expl/service"
	"expl/view"

	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var explainTemplate string = `expl explain mode [option] [sql]

mode:
  simple: explain one sql 
  log:    explain sqls from log file
  log-db: explain sqls from database.

sql:  it's necessary in only "simple" mode.

option:
  -d --database:            string  database. used by any sqls. (using onlh simple mode. not using in table mapping mode)
  -H --host:                string  host used by any sqls.(using onlh simple mode. not using table in mapping mode)
  -u --user:                string  database user used by any sqls.(using onlh simple mode. not using table in mapping mode)
  -p --pass:                string  database password used by any sqls.(using onlh simple mode. not using table in mapping mode)
  -c, --conf:               string  use table mapping. This is efficient for switching hosts and databases by table. 
                                    "EXPL_CONF" environment variable as a default value.
                                      value:
                                        [mapping file path]: database-table mapping file path. default file is ./table_map.yaml.
                                      ex)
                                        -c $GOPATH/bin/table-mapping.yaml
  -l, --log:                string  log file path. This is used in log mode. "EXPL_LOG" environment variable as a default value.
  -f, --format:             string  sql format.
                                      value:
                                        simple (default):  simple is only sql.
                                        official:          sql is offical mysql sql.log's format.
                                        command:           change text using os command. format-cmd option required.
                                      ex)
                                        -f simple
                                        -f official
                                        -f command
  --format-cmd:             string  os command string used only when fomat option is "command"
                                      ex)
                                        --format-cmd "cut -c 14-"
  --filter-select-type:     strings filter results by target select types.
                                      ex)
                                        --filter-select-type simple, subquery
                                        appear results with SIMPLE or SUBQUERY selected-types.
  --filter-no-select-type:  strings filter results without target select types.
  --filter-table:           strings filter results by target tables.
                                      ex)
                                        --filter-table user, group
                                        appear results, table of which is "user" or "group".
  --filter-no-table:        strings filter results without target tables.
  --filter-type:            strings filter results by target types.
                                      ex)
                                        --filter-type index, ref
                                        appear results wich "index" or "ref" types.
  --filter-no-type:         strings filter results without target types.
  --filter-extra:           strings filter results by target taypes.
                                      ex)
                                        --filter-extra filesort, "using where" 
                                        appear results wich "filesort" or "using where" types.
  --filter-no-extra:        strings filter results without target types.

  -U, --update-table-map:           update table-database mapping file before do explain sql. use current database environment.
  -I, --ignore-error:               ignore parse sql error or explain sql error.
  -C, --combine-sql:                This is useful in log or log-db module. combine identical SQL into one.

  --option-file:                    you can use option setting file. with this file you do not have to enter the above optional parameters.
                                    "EXPL_OPTION" environment variable as a default value.
  -v, --verbose:					verbose output.
  -h, --help:                       help. show usage.
`

func validateArgs(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("invalit parameter.")
	}

	mode := args[0]
	switch mode {
	case "simple":
		if len(args) < 2 {
			return fmt.Errorf("sql is none.")
		}
	case "log":
		return nil
	case "log-db":
		return nil
	default:
		return fmt.Errorf("invalid mode. mode mast be (simple, log, log-db).")
	}

	return nil
}

// explainCmd represents the explain command
var explainCmd = &cobra.Command{
	Use:   "explain",
	Short: "A brief description of your command",
	RunE: func(cmd *cobra.Command, args []string) error {

		// validate parameter
		if err := validateArgs(args); err != nil {
			return err
		}

		// mode
		mode := args[0]

		// sql
		sql := ""
		if mode == "simple" {
			sql = args[1]
		}

		optionFilePath, err := cmd.Flags().GetString("option-file")
		if err != nil {
			return err
		}
		if optionFilePath == "" {
			optionFilePath = os.Getenv("EXPL_OPTION")
		}
		if optionFilePath != "" {
			viper.SetConfigFile(optionFilePath)

			if err := viper.ReadInConfig(); err != nil {
				return err
			}
		}

		expOpt := new(model.ExplainOption)

		expOpt.DB = viper.GetString("database")
		expOpt.DBHost = viper.GetString("host")
		expOpt.DBUser = viper.GetString("user")
		expOpt.DBPass = viper.GetString("pass")
		expOpt.Config = viper.GetString("conf")
		if expOpt.Config == "" {
			expOpt.Config = os.Getenv("EXPL_CONF")
		}

		// log file
		logPath := viper.GetString("log")
		if logPath == "" {
			logPath = os.Getenv("EXPL_LOG")
		}

		// format
		format := viper.GetString("format")
		formatCmd := viper.GetString("format-cmd")
		if formatCmd != "" {
			format = string(service.FormatCommand)
		}

		// filter options
		fiOpt := new(model.ExplainFilter)

		fiOpt.SelectType = viper.GetStringSlice("filter-select-type")
		fiOpt.SelectTypeNot = viper.GetStringSlice("filter-no-select-type")
		fiOpt.Table = viper.GetStringSlice("filter-table")
		fiOpt.TableNot = viper.GetStringSlice("filter-no-table")
		fiOpt.Type = viper.GetStringSlice("filter-type")
		fiOpt.TypeNot = viper.GetStringSlice("filter-no-type")
		fiOpt.Extra = viper.GetStringSlice("filter-extra")
		fiOpt.ExtraNot = viper.GetStringSlice("filter-no-extra")

		expOpt.UseTableMap = viper.GetBool("update-table-map")
		if expOpt.DB != "" && expOpt.DBHost != "" && expOpt.DBUser != "" {
			expOpt.UseTableMap = false
		}

		expOpt.NoError = viper.GetBool("ignore-error")
		expOpt.Uniq = viper.GetBool("combine-sql")

		verbose, err := cmd.Flags().GetBool("verbose")
		if err != nil {
			return err
		}
		if verbose {
			view.RenderOptions(expOpt, fiOpt, logPath, format, formatCmd)
		}

		ctx := context.Background()

		switch mode {
		case "simple":
			err = func() error {
				if expOpt.UseTableMap {
					if err = service.ReloadAllTableInfo(ctx, expOpt.Config); err != nil {
						return err
					}

					if err = service.LoadDBInfo(ctx, expOpt.Config); err != nil {
						return err
					}

				} else {
					service.SetDBOne(
						expOpt.DBHost,
						expOpt.DB,
						expOpt.DBUser,
						expOpt.DBPass,
					)
				}

				sql, err = service.GetQueryByFormat(service.FormatType(format), sql, formatCmd)
				if err != nil {
					return err
				}

				list, err := service.Explains(ctx, []string{sql}, expOpt, fiOpt)
				if err == nil {
					view.RenderExplains(list, false)
				}
				return err
			}()
		case "log":
			err = func() error {
				sqls, err := service.LoadQueriesFromLog(ctx, logPath, service.FormatType(format), formatCmd)
				if err != nil {
					return err
				}

				if err = service.ReloadAllTableInfo(ctx, expOpt.Config); err != nil {
					return err
				}

				if err = service.LoadDBInfo(ctx, expOpt.Config); err != nil {
					return err
				}

				list, err := service.Explains(ctx, sqls, expOpt, fiOpt)
				if err == nil {
					view.RenderExplains(list, expOpt.Uniq)
				}
				return err
			}()

		case "log-db":
			err = func() error {
				if err = service.ReloadAllTableInfo(ctx, expOpt.Config); err != nil {
					return err
				}

				if err = service.LoadDBInfo(ctx, expOpt.Config); err != nil {
					return err
				}

				sqls, err := service.LoadQueriesFromDB(ctx)
				if err != nil {
					return err
				}

				list, err := service.Explains(ctx, sqls, expOpt, fiOpt)
				if err == nil {
					view.RenderExplains(list, expOpt.Uniq)
				}
				return err
			}()
		}

		if err != nil {
			fmt.Println("error occured!")
			fmt.Println(service.Message(err))
		}

		return err
	},
}

func init() {
	rootCmd.AddCommand(explainCmd)
	explainCmd.SetUsageTemplate(explainTemplate)
	explainCmd.SetHelpTemplate(explainTemplate)

	//viper.SetConfigType("yml")

	viper.SetDefault("format", "simple")

	viper.SetDefault("filter-select-type", []string{})
	viper.SetDefault("filter-no-select-type", []string{})
	viper.SetDefault("filter-table", []string{})
	viper.SetDefault("filter-no-table", []string{})
	viper.SetDefault("filter-type", []string{})
	viper.SetDefault("filter-no-type", []string{})
	viper.SetDefault("filter-extra", []string{})
	viper.SetDefault("filter-no-extra", []string{})

	viper.SetDefault("update-table-map", true)

	explainCmd.PersistentFlags().StringP("database", "d", "", "database")
	explainCmd.PersistentFlags().StringP("host", "H", "", "host")
	explainCmd.PersistentFlags().StringP("user", "u", "", "database user")
	explainCmd.PersistentFlags().StringP("pass", "p", "", "database password")
	explainCmd.PersistentFlags().StringP("conf", "c", "", "config. which includes database-table mapping file.")
	explainCmd.PersistentFlags().StringP("log", "l", "", "sql log file path.")
	explainCmd.PersistentFlags().StringP("format", "f", "simple", "format of the line.")
	explainCmd.PersistentFlags().StringP("format-cmd", "", "", "os command to update line.")
	explainCmd.PersistentFlags().StringSlice("filter-select-type", []string{}, "filter results by target select types.")
	explainCmd.PersistentFlags().StringSlice("filter-no-select-type", []string{}, "filter results without target select types.")
	explainCmd.PersistentFlags().StringSlice("filter-table", []string{}, "filter results by target tables.")
	explainCmd.PersistentFlags().StringSlice("filter-no-table", []string{}, "filter results without target tables.")
	explainCmd.PersistentFlags().StringSlice("filter-type", []string{}, "filter results by target types.")
	explainCmd.PersistentFlags().StringSlice("filter-no-type", []string{}, "filter results without target types.")
	explainCmd.PersistentFlags().StringSlice("filter-extra", []string{}, "strings filter results by target taypes.")
	explainCmd.PersistentFlags().StringSlice("filter-no-extra", []string{}, "filter results without target types.")
	explainCmd.PersistentFlags().BoolP("update-table-map", "U", true, "update table-database mapping file before do explain sql. use current database environment.")
	explainCmd.PersistentFlags().BoolP("ignore-error", "I", false, "ignore sql error.")
	explainCmd.PersistentFlags().BoolP("combine-sql", "C", false, "This is useful in log or log-db module. combine identical SQL into one.")

	explainCmd.Flags().StringP("option-file", "", "", "option yaml file.")
	explainCmd.Flags().BoolP("verbose", "v", false, "verbose output.")

	viper.BindPFlag("database", explainCmd.PersistentFlags().Lookup("database"))
	viper.BindPFlag("database", explainCmd.PersistentFlags().ShorthandLookup("d"))
	viper.BindPFlag("host", explainCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("host", explainCmd.PersistentFlags().ShorthandLookup("H"))
	viper.BindPFlag("user", explainCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("user", explainCmd.PersistentFlags().ShorthandLookup("u"))
	viper.BindPFlag("pass", explainCmd.PersistentFlags().Lookup("pass"))
	viper.BindPFlag("pass", explainCmd.PersistentFlags().ShorthandLookup("p"))
	viper.BindPFlag("conf", explainCmd.PersistentFlags().Lookup("conf"))
	viper.BindPFlag("conf", explainCmd.PersistentFlags().ShorthandLookup("c"))
	viper.BindPFlag("log", explainCmd.PersistentFlags().Lookup("log"))
	viper.BindPFlag("log", explainCmd.PersistentFlags().ShorthandLookup("l"))
	viper.BindPFlag("format", explainCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("format", explainCmd.PersistentFlags().ShorthandLookup("f"))
	viper.BindPFlag("format-cmd", explainCmd.PersistentFlags().Lookup("format-cmd"))
	viper.BindPFlag("filter-select-type", explainCmd.PersistentFlags().Lookup("filter-select-type"))
	viper.BindPFlag("filter-no-select-type", explainCmd.PersistentFlags().Lookup("filter-no-select-type"))
	viper.BindPFlag("filter-table", explainCmd.PersistentFlags().Lookup("filter-table"))
	viper.BindPFlag("filter-no-table", explainCmd.PersistentFlags().Lookup("filter-no-table"))
	viper.BindPFlag("filter-type", explainCmd.PersistentFlags().Lookup("filter-type"))
	viper.BindPFlag("filter-no-type", explainCmd.PersistentFlags().Lookup("filter-no-type"))
	viper.BindPFlag("filter-extra", explainCmd.PersistentFlags().Lookup("filter-extra"))
	viper.BindPFlag("filter-no-extra", explainCmd.PersistentFlags().Lookup("filter-no-extra"))
	viper.BindPFlag("update-table-map", explainCmd.PersistentFlags().Lookup("update-table-map"))
	viper.BindPFlag("update-table-map", explainCmd.PersistentFlags().ShorthandLookup("U"))
	viper.BindPFlag("ignore-error", explainCmd.PersistentFlags().Lookup("ignore-error"))
	viper.BindPFlag("ignore-error", explainCmd.PersistentFlags().ShorthandLookup("I"))
	viper.BindPFlag("combine-sql", explainCmd.PersistentFlags().Lookup("combine-sql"))
	viper.BindPFlag("combine-sql", explainCmd.PersistentFlags().ShorthandLookup("C"))
}
