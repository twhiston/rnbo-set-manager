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
	"slices"
	"time"

	"github.com/twhiston/rnbo-set-manager/rnbo"

	"github.com/antandros/go-dpkg"
	"github.com/antandros/go-pkgparser/model"
	_ "github.com/glebarez/go-sqlite"
	"github.com/spf13/cobra"
)

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "export set data from the rnbo sqlite database",
	Long: `Export everything that you need to recreate a set from the rnbo database
	It is important to note that this will export data from the patcher table, so it expects
	that when you import patchers with the correct ID's will already exist in the database`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Export Set")
		if len(args) != 1 {
			log.Fatal("Must specify set name to export as first argument to command")
		}
		rnboSet := args[0]
		fmt.Println("RNBO Set: " + rnboSet)

		rnboVer, _ := cmd.Flags().GetString("rnbo-version")

		if rnboVer == "" {
			rnboVer = determineRnboVersion()
			if rnboVer == "" {
				log.Fatal("Could not determine version of rnbooscquery, are you running this on an rpi?")
			}
		}
		fmt.Println("RNBO version: " + rnboVer)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		rnboDB, _ := cmd.Flags().GetString("db")
		if rnboDB == "" {
			//~/Documents/rnbo/oscqueryrunner.sqlite
			rnboDB = filepath.Join(homeDir, "Documents", "rnbo", "oscqueryrunner.sqlite")
		}
		fmt.Println("RNBO db: " + rnboDB)

		exportDir, _ := cmd.Flags().GetString("dir")
		if exportDir == "" {
			exportDir = filepath.Join(homeDir, "Documents", "rnbo-set-manager-data")
		}
		fmt.Println("Export Base Dir: " + exportDir)

		timeNow := time.Now().Format("20060102-150405")
		exportDir = filepath.Join(exportDir, rnboSet, timeNow)
		fmt.Println("Export Full Dir: " + exportDir)

		// connect
		db, err := sql.Open("sqlite", rnboDB)
		if err != nil {
			log.Fatal(err)
		}

		set := getSet(db, rnboSet, rnboVer)
		setCons := getSetCons(db, set.Id)
		setPis := getSetPI(db, set.Id)
		setPs := getSetPresets(db, set.Id)

		setData, err := json.Marshal(set)
		if err != nil {
			fmt.Println(err)
			return
		}

		setConsData, err := json.Marshal(setCons)
		if err != nil {
			fmt.Println(err)
			return
		}

		setPiData, err := json.Marshal(setPis)
		if err != nil {
			fmt.Println(err)
			return
		}

		setPsData, err := json.Marshal(setPs)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = os.MkdirAll(exportDir, os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}

		writeFile(setData, exportDir, rnboSet, "set")
		writeFile(setConsData, exportDir, rnboSet, "connections")
		writeFile(setPiData, exportDir, rnboSet, "patcher_instances")
		writeFile(setPsData, exportDir, rnboSet, "presets")

		fmt.Println("Export completed")

	},
}

func determineRnboVersion() string {
	version := ""

	packages, err := dpkg.GetPackages()
	if err != nil {
		log.Fatal(err)
	}

	idx := slices.IndexFunc(packages, func(c model.Package) bool { return c.PackageName == "rnbooscquery" })
	if idx != -1 {
		version = packages[idx].Version
	}
	return version
}

func getSet(db *sql.DB, setId string, version string) rnbo.Set {
	var set rnbo.Set
	if err := db.QueryRow("SELECT id,name,filename,runner_rnbo_version,created_at,meta FROM sets where name = ? AND runner_rnbo_version = ?",
		setId, version).Scan(&set.Id, &set.Name, &set.Filename, &set.Runner_rnbo_version, &set.Created_at, &set.Meta); err != nil {
		fmt.Println("Issue getting set")
		log.Fatal(err)
	}

	return set
}

func getSetCons(db *sql.DB, id int) []rnbo.SetConnection {
	var setCons []rnbo.SetConnection
	res, err := db.Query("Select * FROM sets_connections where set_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	for res.Next() {
		setC := &rnbo.SetConnection{}

		err = res.Scan(
			&setC.Id,
			&setC.Set_Id,
			&setC.Source_name,
			&setC.Source_instance_index,
			&setC.Source_port_name,
			&setC.Sink_name,
			&setC.Sink_instance_index,
			&setC.Sink_port_name,
		)

		if err == nil {
			setCons = append(setCons, *setC)
		} else {
			log.Fatal(err)
		}
	}
	return setCons
}

func getSetPI(db *sql.DB, id int) []rnbo.SetPatcherInstance {
	var setPInst []rnbo.SetPatcherInstance
	res, err := db.Query("Select * FROM sets_patcher_instances where set_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	for res.Next() {
		setPI := &rnbo.SetPatcherInstance{}

		err = res.Scan(
			&setPI.Id,
			&setPI.Patcher_id,
			&setPI.Set_id,
			&setPI.Set_instance_index,
			&setPI.Config,
		)

		if err == nil {
			setPInst = append(setPInst, *setPI)
		} else {
			log.Fatal(err)
		}
	}
	return setPInst
}

func getSetPresets(db *sql.DB, id int) []rnbo.SetPreset {
	var setPs []rnbo.SetPreset
	res, err := db.Query("Select * FROM sets_presets where set_id = ?", id)
	if err != nil {
		log.Fatal(err)
	}
	for res.Next() {
		setP := &rnbo.SetPreset{}

		err = res.Scan(
			&setP.Id,
			&setP.Patcher_id,
			&setP.Set_id,
			&setP.Set_instance_index,
			&setP.Name,
			&setP.Content,
			&setP.Initial,
			&setP.Created_at,
			&setP.Updated_at,
		)

		if err == nil {
			setPs = append(setPs, *setP)
		} else {
			log.Fatal(err)
		}
	}
	return setPs
}

func writeFile(data []byte, path string, setName string, uid string) {
	fpath := filepath.Join(path, setName+"_"+uid+".json")
	fc, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer fc.Close()
	fc.Write(data)
}

func init() {
	rootCmd.AddCommand(exportCmd)

	// Here you will define your flags and configuration settings.
	exportCmd.Flags().String("rnbo-version", "", "Set the rnbo runner version, leave blank to autodetect")
}
