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

// USAGE: GITHUB_AUTH_TOKEN=`gh auth token` go run update_packages.go > ../packages

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"os"

	"github.com/google/go-github/v54/github"
)

func getRepos(ctx context.Context, client *github.Client, organizationName string) ([]*github.Repository, error) {
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

	// Filter archived repos
	var active []*github.Repository
	for _, repo := range allRepos {
		if *repo.Archived == false {
			active = append(active, repo)
		}
	}

	//fmt.Printf("total repos found: %d\n", len(allRepos))
	return active, nil
}

func main() {

	ctx := context.Background()

	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		log.Fatal("Unauthorized: No token present in GITHUB_AUTH_TOKEN")
	}
	client := github.NewTokenClient(ctx, token)

	repos, err := getRepos(ctx, client, "solus-packages")
	if err != nil {
		log.Fatalf("Failed to get repos, reason: %s\n", err)
	}

	var repoNames []string
	for _, repo := range repos {
		if *repo.Name != ".github" {
			repoNames = append(repoNames, *repo.Name)
		}
	}

	sort.Strings(repoNames)

	for _, name := range repoNames {
		fmt.Println(name)
	}
}
