package main

import (
	azurefileshare "github.com/vpavlin/azure-file-share/cmd/azurefileshare"
)

var Version = "v0.0.0"

func main() {
	azurefileshare.Execute()
}
