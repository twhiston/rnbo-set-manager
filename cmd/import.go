/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import set data into rnbo runner database",
	Long: `This command will import a previously exported Set to the rnbo database.
	Note that it does not recreate any of the compiled objects, or ensure their ID's are valid
	So it expects that you have not destroyed these objects etc.
	By default the command will import your Set with a NEW set ID. You can tell it to use the original ID by setting the original-id flag.
	Use this at your own risk!`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Import Set")
		if len(args) != 1 {
			log.Fatal("Must specify set name to import as first argument to command")
		}
		rnboSet := args[0]
		fmt.Println("RNBO Set: " + rnboSet)

		overwrite, _ := cmd.Flags().GetBool("original-id")
		fmt.Println("Import with original Id: ", overwrite)

		//timestamp, _ := cmd.Flags().GetString("timestamp")
		//If timestamp is blank then look in the folder and find the appropriate timestamp

	},
}

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	importCmd.Flags().Bool("original-id", false, "Use the original set ID instead of creating a new one. Use at your own risk! Probably a good idea to backup your db first!")
	importCmd.Flags().String("timestamp", "", "Specify a specifically timestamped export of the set to import, if blank will get most recent")
}
