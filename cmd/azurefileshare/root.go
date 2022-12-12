package azurefileshare

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var version = "v0.0.0"

var rootCmd = &cobra.Command{
	Use:     "azurefileshare",
	Version: version,
	PreRun: func(cmd *cobra.Command, args []string) {
		ef, err := cmd.PersistentFlags().GetString("env-file")
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Using Env File ", ef)
		err = godotenv.Load(ef)
		if err != nil {
			log.Fatal("Error loading .env file")
		}

	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		log.Fatalln(err)
	}
}

func init() {
	rootCmd.PersistentFlags().String("env-file", ".env", "Path to an dotenv file")
}
