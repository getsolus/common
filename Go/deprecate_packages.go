/*
 * Copyright 2023 Solus Project <copyright@getsol.us>
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http: *www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// USAGE: GITHUB_AUTH_TOKEN=`gh auth token` go run deprecate_packages.go

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/google/go-github/v54/github"
)

func getAllRepos(ctx context.Context, client *github.Client, organizationName string) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		Type: "public",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var rateErr *github.RateLimitError
	var abuseErr *github.AbuseRateLimitError

	// get all pages of results
	var allRepos []*github.Repository
	for {
		repos, resp, err := client.Repositories.ListByOrg(ctx, organizationName, opt)

		if errors.As(err, &rateErr) {
			log.Fatalln("hit rate limit")
		}
		if errors.As(err, &abuseErr) {
			log.Fatalln("hit secondary rate limit")
		}
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		log.Printf("Getting repos, page %d\n", opt.Page)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	//fmt.Printf("total repos found: %d\n", len(allRepos))
	return allRepos, nil
}

func contains[T comparable](s []T, e T) bool {
    for _, v := range s {
        if v == e {
            return true
        }
    }
    return false
}

func main() {
	ctx := context.Background()

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present in GITHUB_AUTH_TOKEN")
	}
	client := github.NewTokenClient(ctx, token)

	repos, err := getAllRepos(ctx, client, "solus-packages")
	if err != nil {
		log.Fatalf("Failed to get repos, reason: %s\n", err)
	}

	// FIXME: This is ugly but we need to run from the folder with the go.mod file
	rootDir := "../../"

	entries, err := os.ReadDir(rootDir)
	if err != nil {
		panic(err.Error())
	}
	stdin := bufio.NewReader(os.Stdin)
	remove := false
	done := false
	fmt.Printf("The following repos (directories) will be removed:\n\n")
	var removals []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		for _, repo := range repos {
			if *repo.Archived == true {
				if string(entry.Name()) == string(*repo.Name) {
					fmt.Println(entry.Name())
					removals = append(removals, entry.Name())
				}
			}
		}
	}

	fmt.Println()
	for !done {
		fmt.Println("Would you like to remove of all archived repos? (yes/no)")
		ans, err := stdin.ReadString('\n')
		if err != nil {
			panic(err.Error())
		}
		switch ans {
		case "yes\n":
			remove = true
			done = true
		case "no\n":
			remove = false
			done = true
		default:
			fmt.Printf("'%s' is not a valid answer", ans)
		}
	}
	if !remove {
		os.Exit(0)
	}
	for _, dir := range removals {
		fmt.Printf("Removing repository '%s'...", dir)
		if err := os.RemoveAll(rootDir + dir); err != nil {
			fmt.Printf("FAILED: %s\n", err.Error())
		} else {
			fmt.Println("DONE")
		}
	}
}
