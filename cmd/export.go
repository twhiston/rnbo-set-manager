/*
Copyright © 2024 Tom Whiston tom@tomwhiston.com
*/
package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/antandros/go-dpkg"
	_ "github.com/glebarez/go-sqlite"
	"github.com/spf13/cobra"
)

type RnboSet struct {
	Id                  int
	Name                string
	Filename            string
	Runner_rnbo_version string
	Created_at          string
	Meta                string
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export set data from the rnbo sqlite database",
	Long:  `Export everything that you need to recreate a set from the rnbo database`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Export")
		if len(args) != 1 {
			log.Fatal("Must specify set name to export as first argument to command")
		}
		rnboSet := args[0]
		fmt.Println("RNBO Set: " + rnboSet)

		rnboVer, _ := cmd.Flags().GetString("rnbo-version")

		if rnboVer == "" {
			rnboVer = determineRnboVersion()
		}
		fmt.Println("RNBO version: " + rnboVer)

		rnboDB, _ := cmd.Flags().GetString("db")
		fmt.Println("RNBO db: " + rnboDB)

		exportDir, _ := cmd.Flags().GetString("dir")
		fmt.Println("Export Dir: " + exportDir)

		// connect
		db, err := sql.Open("sqlite", rnboDB)
		if err != nil {
			log.Fatal(err)
		}

		var set RnboSet
		if err := db.QueryRow("SELECT id,name,filename,runner_rnbo_version,created_at,meta FROM sets where name = ? AND runner_rnbo_version = ?",
			rnboSet, rnboVer).Scan(&set.Id, &set.Name, &set.Filename, &set.Runner_rnbo_version, &set.Created_at, &set.Meta); err != nil {
			fmt.Println("Issue getting set")
			log.Fatal(err)
		}

		exportData, err := json.Marshal(set)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = os.MkdirAll(exportDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		path := filepath.Join(exportDir, rnboSet+"_exported_"+time.Now().Format("20060102-150405")+".json")
		f, err := os.Create(path)
		if err != nil {
			fmt.Println(err)
			return
		}
		defer f.Close()
		f.Write(exportData)
	},
}

func determineRnboVersion() string {
	packages, err := dpkg.GetPackages()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(err)
	resp, err := json.MarshalIndent(packages, "", "\t")
	fmt.Println(err)
	fmt.Println(string(resp))
	return "1.3.3-alpha.0"
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// exportCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// exportCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	exportCmd.Flags().String("rnbo-version", "", "Set the rnbo runner version, leave blank to autodetect")
	exportCmd.Flags().String("db", "~/Documents/rnbo/oscqueryrunner.sqlite", "Specify the location to the db")
	exportCmd.Flags().String("dir", "~/Documents/rnbo-set-manager", "Specify save file location")
}
