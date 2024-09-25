/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
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

	"github.com/spf13/cobra"
	"github.com/twhiston/rnbo-set-manager/rnbo"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import set data into rnbo runner database",
	Long: `This command will import a previously exported Set and associated data to the rnbo database.
Note that it does not recreate any of the compiled objects, or ensure their ID's are valid
By default the command will import your Set with a NEW set Name and ID.
The new set name will be:
		set.Name + "_" + importTimestamp + "_restored_" + nowTimestamp
or use --name flag to set something.

	IT IS STRONGLY RECOMMENDED TO BACKUP YOUR DATABASE BEFORE IMPORTING ANY SETS`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Import Set")
		if len(args) != 1 {
			log.Fatal("Must specify set name to import as first argument to command")
		}
		rnboSet := args[0]
		fmt.Println("RNBO Set: " + rnboSet)

		homeDir, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}

		rnboDB, _ := cmd.Flags().GetString("db")
		if rnboDB == "" {
			rnboDB = filepath.Join(homeDir, "Documents", "rnbo", "oscqueryrunner.sqlite")
		}
		fmt.Println("RNBO db: " + rnboDB)

		// connect to db or fail
		db, err := sql.Open("sqlite", rnboDB)
		if err != nil {
			log.Fatal(err)
		}

		importDir, _ := cmd.Flags().GetString("dir")
		if importDir == "" {
			importDir = filepath.Join(homeDir, "Documents", "rnbo-set-manager-data")
		}
		fmt.Println("Import Base Dir: " + importDir)
		importDir = filepath.Join(importDir, rnboSet)

		//TODO: implement functionality around this
		// overwrite, _ := cmd.Flags().GetBool("original-id")
		// fmt.Println("Import with original Id: ", overwrite)

		timestamp, _ := cmd.Flags().GetString("timestamp")
		if timestamp == "" {
			//If timestamp is blank then look in the folder and find the appropriate timestamp
			timestamp = findMostRecentBackup(importDir)
		}

		fullImportPath := filepath.Join(importDir, timestamp)
		if !exists(fullImportPath) {
			log.Fatal("Set and/or timestamp does not exist")
		}

		//Get all the data files and unmarshall them into appropriate types
		rawSetData := readFile(fullImportPath, getFileName(rnboSet, "set"))
		var set rnbo.Set
		err = json.Unmarshal(rawSetData, &set)
		if err != nil {
			log.Fatal(err)
		}

		rawPiData := readFile(fullImportPath, getFileName(rnboSet, "patcher_instances"))
		var pis []rnbo.SetPatcherInstance
		err = json.Unmarshal(rawPiData, &pis)
		if err != nil {
			log.Fatal(err)
		}

		rawConnectionData := readFile(fullImportPath, getFileName(rnboSet, "connections"))
		var cons []rnbo.SetConnection
		err = json.Unmarshal(rawConnectionData, &cons)
		if err != nil {
			log.Fatal(err)
		}

		rawPData := readFile(fullImportPath, getFileName(rnboSet, "presets"))
		var ps []rnbo.SetPreset
		err = json.Unmarshal(rawPData, &ps)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Successfully loaded all data from export")

		//Import all the data into the db
		//Importing will create a new entry in the database
		newSetName, _ := cmd.Flags().GetString("name")
		if newSetName == "" {
			timeNow := time.Now().Format("20060102-150405")
			newSetName = set.Name + "_" + timestamp + "_restored_" + timeNow
		}
		fmt.Println("New Set Name:", newSetName)

		//Import the set and get it's ID
		result, err := db.Exec("INSERT INTO sets(name, filename, runner_rnbo_version, meta) VALUES (?, ?, ?, ?)", newSetName, set.Filename, set.Runner_rnbo_version, set.Meta)
		if err != nil {
			log.Fatal(err)
		}
		newSetID, err := result.LastInsertId()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("New Set ID:", newSetID)

		//Import patcher instances
		//Here we need to change the set_id only
		for _, pInstance := range pis {
			_, err = db.Exec("INSERT INTO sets_patcher_instances(patcher_id, set_id, set_instance_index, config) VALUES (?, ?, ?, ?)", pInstance.Patcher_id, newSetID, pInstance.Set_instance_index, pInstance.Config)
			if err != nil {
				log.Printf("WARNING - INCOMPLETE IMPORT, CONSIDER RESTORING A BACKUP")
				log.Fatal(err)
			}
		}

		//Import the connections, we also only need to alter the set_id here
		for _, connection := range cons {
			_, err = db.Exec("INSERT INTO sets_connections(set_id, source_name, source_instance_index, source_port_name, sink_name, sink_instance_index, sink_port_name) VALUES (?, ?, ?, ?, ?, ?, ?)", newSetID, connection.Source_name, connection.Source_instance_index, connection.Source_port_name, connection.Sink_name, connection.Sink_instance_index, connection.Sink_port_name)
			if err != nil {
				log.Printf("WARNING - INCOMPLETE IMPORT, CONSIDER RESTORING A BACKUP")
				log.Fatal(err)
			}
		}

		//Import set presets
		for _, preset := range ps {
			_, err = db.Exec("INSERT INTO sets_presets(patcher_id, set_id, set_instance_index, name, content, initial) VALUES (?, ?, ?, ?, ?, ?)", preset.Patcher_id, newSetID, preset.Set_instance_index, preset.Name, preset.Content, preset.Initial)
			if err != nil {
				log.Printf("WARNING - INCOMPLETE IMPORT, CONSIDER RESTORING A BACKUP")
				log.Fatal(err)
			}
		}

		fmt.Println("Importing Set Complete")
		fmt.Println("Reboot your Pi to see imported set")

	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.
	//importCmd.Flags().Bool("original-id", false, "Use the original set ID instead of creating a new one. NOT CURRENTLY IMPLEMENTED")
	importCmd.Flags().String("timestamp", "", "Specify a specifically timestamped export of the set to import, if blank will get most recent")
	importCmd.Flags().String("name", "", "Specify a specific set name for the import, if blank will be \"set.Name + \"_\" + importTimestamp + \"_restored_\" + nowTimestamp")
}

func getFileName(setName string, uid string) string {
	return setName + "_" + uid + ".json"
}

func findMostRecentBackup(dir string) string {
	files, _ := os.ReadDir(dir)
	var newestFolder string
	var newestTime int64 = 0
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		fi, err := f.Info()
		if err != nil {
			log.Println(err)
			continue
		}
		currTime := fi.ModTime().Unix()
		if currTime > newestTime {
			newestTime = currTime
			newestFolder = f.Name()
		}
	}
	return newestFolder
}

func exists(filepath string) bool {
	if _, err := os.Stat(filepath); err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func readFile(targetDir string, pattern string) []byte {

	var data []byte

	matches, err := filepath.Glob(filepath.Join(targetDir, pattern))

	if err != nil {
		log.Fatal(err)
	}

	if len(matches) != 0 {
		data, err = os.ReadFile(matches[0])
		if err != nil {
			log.Fatal(err)
		}
	}
	return data
}
