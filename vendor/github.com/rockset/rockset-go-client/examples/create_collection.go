package main

import (
	"fmt"
	"os"

	apiclient "github.com/rockset/rockset-go-client"
	models "github.com/rockset/rockset-go-client/lib/go"
)

func main() {
	apiKey := os.Getenv("ROCKSET_APIKEY")
	apiServer := os.Getenv("ROCKSET_APISERVER")

	// create the API client
	client := apiclient.Client(apiKey, apiServer)

	// create collection
	cinfo := models.CreateCollectionRequest{
		Name:        "my-first-collection",
		Description: "my first collection",
	}

	res, _, err := client.Collection.Create(cinfo)
	if err != nil {
		fmt.Printf("error: %s\n", err)
		return
	}

	res.PrintResponse()
}
