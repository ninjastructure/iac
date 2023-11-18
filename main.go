package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/go-github/v56/github"
)

type Iac struct {
	AwsAccessKey *Secret
	AwsSecretKey *Secret
	PulumiToken  *Secret
}

func (m *Iac) WithCredentials(pulumiToken, awsAccessKey, awsSecretKey *Secret) *Iac {
	m.PulumiToken = pulumiToken
	m.AwsAccessKey = awsAccessKey
	m.AwsSecretKey = awsSecretKey
	return m
}

func (m *Iac) Preview(ctx context.Context, src *Directory, stack string, githubToken *Secret, githubRef, githubRepo string) error {
	diff, err := dag.Pulumi().
		WithAwsCredentials(m.AwsAccessKey, m.AwsSecretKey).
		WithPulumiToken(m.PulumiToken).
		Preview(ctx, src, stack)
	if err != nil {
		return err
	}

	// On pull requests GITHUB_REF has the format: `refs/pull/:prNumber/merge`
	pr, err := strconv.Atoi(strings.Split(githubRef, "/")[2])
	if err != nil {
		return fmt.Errorf("githubRef did not have the correct format, expected: refs/pull/:prNumber/merge. Got: %s", githubRef)
	}

	token, err := githubToken.Plaintext(ctx)
	if err != nil {
		return err
	}

	repo := strings.Split(githubRepo, "/")
	return postComment(ctx, diff, token, repo[0], repo[1], pr)
}

func (m *Iac) Up(ctx context.Context, src *Directory, stack string) (string, error) {
	return dag.Pulumi().
		WithAwsCredentials(m.AwsAccessKey, m.AwsSecretKey).
		WithPulumiToken(m.PulumiToken).
		Up(ctx, src, stack)
}

func postComment(ctx context.Context, content, githubToken, owner, repo string, pr int) error {
	body := fmt.Sprintf("```\n%s\n```", content)
	client := github.NewClient(nil).WithAuthToken(githubToken)
	_, _, err := client.Issues.CreateComment(ctx, owner, repo, pr, &github.IssueComment{
		Body: &body,
	})
	return err
}
