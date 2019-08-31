package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/google/go-github/v24/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

func main() {

	err := godotenv.Load("./docker/.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Check if GITHUBTOKEN is present in env
	token := os.Getenv("GITHUBTOKEN")
	if token == "" {
		log.Fatalf("No github token found")
	}

	//Background Context
	ctx := context.Background()

	tokenSource := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, tokenSource)

	// Github Client
	client := github.NewClient(tc)

	// Query Parameters for Tags API
	listOptions := &github.ListOptions{
		Page:    1,
		PerPage: 3,
	}

	// Get All Tags for Swift Repository
	tags, _, err := client.Repositories.ListTags(ctx, "apple", "swift", listOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Loop through all tags.
	for _, element := range tags {

		// Check if tag name is not type-name-lookup-fail, for some reason api responds with this as first element
		if strings.Contains(*element.Name, "DEVELOPMENT") {

			// Set Swift Version
			os.Setenv("SWIFT_VERSION", *element.Name)
			log.Println(os.Getenv("SWIFT_VERSION"))

			// Get and Set Swift Version in Enviroment File.
			env, err := godotenv.Unmarshal("SWIFT_VERSION=" + *element.Name)
			if err != nil {
				log.Fatal(err)
			}
			godotenv.Write(env, "./docker/.env")

			// Read Dockerfile
			dockerfile, err := ioutil.ReadFile("./docker/Dockerfile")
			if err != nil {
				log.Fatal(err)
			}
			strFile := string(dockerfile)

			// check if Dockerfile contain Swift Version Enviroment Variable
			if strings.Contains(strFile, "ARG SWIFT_VERSION") {

				// Get and Set SWIFT_VERSION
				index := strings.Index(strFile, "ARG SWIFT_VERSION")
				newArg := "ARG SWIFT_VERSION=" + *element.Name
				newDockerfile := strings.Replace(strFile, "ARG SWIFT_VERSION", newArg, index)

				// Save Dockerfile with New Swift Version
				f, err := os.OpenFile("./docker/Dockerfile", os.O_WRONLY, 0644)
				if err != nil {
					log.Fatal(err)
				}
				if _, err := f.Write([]byte(newDockerfile)); err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}
