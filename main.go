/*
gh-user-recon
https://github.com/x1sec

See: LICENSE.md
*/
package main

import (
	"fmt"
	"github.com/google/go-github/github"
	"github.com/jedib0t/go-pretty/v6/table"
	"golang.org/x/oauth2"
	"strings"
	"context"
	"flag"
	"os"
)

type User struct {
	Name string
	Email string
}

func main() {
	user := flag.String("u", "", "Github username")
	token := flag.String("t", "", "Github token")
	flag.Parse()
	if *user == "" {
		fmt.Println("[-] No username specified, exiting")
		os.Exit(1)
	}
	if *token == "" {
		*token = os.Getenv("GITHUB_TOKEN")
		if *token == "" {
			fmt.Println("[-] No Github token specified. Recommended to set one")
		}
	}
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: *token},
	)
	tc := oauth2.NewClient(ctx, ts)

	if *token == "" {
		tc = nil
	}
	client := github.NewClient(tc)	
	//orgs, _, err := client.Organizations.List(context.Background(), "x1sec", nil)
	fmt.Printf("[+] Fetching ...\n")
	repos, _, err := client.Repositories.List(ctx, *user, nil)

	results := make(map[string][]User)
	if err != nil {
		panic(err)
	}
	fmt.Printf("[+] enumerating %d repositories for user '%s'\n", len(repos), *user)
	for _,r := range repos {
		repoName := *r.Name
		results[repoName] = EnumCommits(client, *user, *r.Name)
	}
	PrintTable(results)
}
func EnumCommits(client *github.Client, user string, repoName string) []User {
	commits, response, err := client.Repositories.ListCommits(context.Background(), user, repoName, nil)
    if err != nil && response.StatusCode != 200 {
        panic(err)
    }

	unique := make(map[string]bool)
	var users []User
    for _, commit := range commits {
		email := *commit.Commit.Author.Email
		if strings.Contains(email, "@") {
			s := strings.Split(email, "@")
			if s[1] == "users.noreply.github.com" {
				continue
			}
		}
		name := *commit.Commit.Author.Name
		if _, value := unique[email+name]; !value {
			unique[email+name] = true
			users = append(users, User{Email: email, Name: name})
		} 
    }
	return users
}

func PrintTable(results map[string][]User) {

	for r := range results {
		if len(results[r]) == 0 {
			continue
		}
		t := table.NewWriter()
		t.SetStyle(table.StyleColoredBright)
		t.SetTitle(r)
		t.SetColumnConfigs([]table.ColumnConfig{
			{Number:1,WidthMin:15},
			{Number:2,WidthMin:30},
		})
		t.AppendHeader(table.Row{"Name", "Email"})
		for _, u := range results[r] {
			t.AppendRow(table.Row{u.Name, u.Email})
		}
		fmt.Println(t.Render())
		fmt.Println("\n")
	}
}
