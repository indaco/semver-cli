package changeloggenerator

import (
	"testing"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name         string
		url          string
		wantProvider string
		wantHost     string
		wantOwner    string
		wantRepo     string
		wantErr      bool
	}{
		// GitHub
		{
			name:         "GitHub SSH format",
			url:          "git@github.com:indaco/semver-cli.git",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "indaco",
			wantRepo:     "semver-cli",
		},
		{
			name:         "GitHub SSH format without .git",
			url:          "git@github.com:owner/repo",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "owner",
			wantRepo:     "repo",
		},
		{
			name:         "GitHub HTTPS format",
			url:          "https://github.com/indaco/semver-cli.git",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "indaco",
			wantRepo:     "semver-cli",
		},
		{
			name:         "GitHub HTTPS format without .git",
			url:          "https://github.com/owner/repo",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "owner",
			wantRepo:     "repo",
		},
		{
			name:         "GitHub Git protocol",
			url:          "git://github.com/owner/repo.git",
			wantProvider: "github",
			wantHost:     "github.com",
			wantOwner:    "owner",
			wantRepo:     "repo",
		},
		// GitLab
		{
			name:         "GitLab SSH format",
			url:          "git@gitlab.com:mygroup/myproject.git",
			wantProvider: "gitlab",
			wantHost:     "gitlab.com",
			wantOwner:    "mygroup",
			wantRepo:     "myproject",
		},
		{
			name:         "GitLab HTTPS format",
			url:          "https://gitlab.com/mygroup/myproject.git",
			wantProvider: "gitlab",
			wantHost:     "gitlab.com",
			wantOwner:    "mygroup",
			wantRepo:     "myproject",
		},
		// Codeberg
		{
			name:         "Codeberg SSH format",
			url:          "git@codeberg.org:user/project.git",
			wantProvider: "codeberg",
			wantHost:     "codeberg.org",
			wantOwner:    "user",
			wantRepo:     "project",
		},
		{
			name:         "Codeberg HTTPS format",
			url:          "https://codeberg.org/user/project",
			wantProvider: "codeberg",
			wantHost:     "codeberg.org",
			wantOwner:    "user",
			wantRepo:     "project",
		},
		// Bitbucket
		{
			name:         "Bitbucket SSH format",
			url:          "git@bitbucket.org:team/repo.git",
			wantProvider: "bitbucket",
			wantHost:     "bitbucket.org",
			wantOwner:    "team",
			wantRepo:     "repo",
		},
		{
			name:         "Bitbucket HTTPS format",
			url:          "https://bitbucket.org/team/repo.git",
			wantProvider: "bitbucket",
			wantHost:     "bitbucket.org",
			wantOwner:    "team",
			wantRepo:     "repo",
		},
		// Custom/self-hosted
		{
			name:         "Self-hosted GitLab SSH",
			url:          "git@git.company.com:team/project.git",
			wantProvider: "custom",
			wantHost:     "git.company.com",
			wantOwner:    "team",
			wantRepo:     "project",
		},
		{
			name:         "Self-hosted Gitea HTTPS",
			url:          "https://gitea.myserver.io/user/repo",
			wantProvider: "custom",
			wantHost:     "gitea.myserver.io",
			wantOwner:    "user",
			wantRepo:     "repo",
		},
		// Error cases
		{
			name:    "Invalid URL",
			url:     "not-a-valid-url",
			wantErr: true,
		},
		{
			name:    "Local path",
			url:     "/path/to/repo",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRemoteURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}
			if got.Host != tt.wantHost {
				t.Errorf("Host = %q, want %q", got.Host, tt.wantHost)
			}
			if got.Owner != tt.wantOwner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.wantOwner)
			}
			if got.Repo != tt.wantRepo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.wantRepo)
			}
		})
	}
}

func TestExtractUsername(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		authorName string
		wantUser   string
		wantHost   string
	}{
		{
			name:       "GitHub noreply with ID",
			email:      "12345+testuser@users.noreply.github.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "github.com",
		},
		{
			name:       "GitHub noreply without ID",
			email:      "testuser@users.noreply.github.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "github.com",
		},
		{
			name:       "GitLab noreply",
			email:      "testuser@noreply.gitlab.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "gitlab.com",
		},
		{
			name:       "Codeberg noreply",
			email:      "myuser@noreply.codeberg.org",
			authorName: "My User",
			wantUser:   "myuser",
			wantHost:   "codeberg.org",
		},
		{
			name:       "Regular email - fallback to author name",
			email:      "test@example.com",
			authorName: "Test User",
			wantUser:   "testuser",
			wantHost:   "",
		},
		{
			name:       "Single name author",
			email:      "user@example.com",
			authorName: "Developer",
			wantUser:   "developer",
			wantHost:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, gotHost := extractUsername(tt.email, tt.authorName)
			if gotUser != tt.wantUser {
				t.Errorf("username = %q, want %q", gotUser, tt.wantUser)
			}
			if gotHost != tt.wantHost {
				t.Errorf("host = %q, want %q", gotHost, tt.wantHost)
			}
		})
	}
}

func TestGetContributors(t *testing.T) {
	commits := []CommitInfo{
		{Author: "Alice", AuthorEmail: "alice@example.com"},
		{Author: "Bob", AuthorEmail: "bob@example.com"},
		{Author: "Alice", AuthorEmail: "alice@example.com"}, // Duplicate
		{Author: "Charlie", AuthorEmail: "charlie@users.noreply.github.com"},
	}

	contributors := getContributors(commits)

	if len(contributors) != 3 {
		t.Fatalf("expected 3 unique contributors, got %d", len(contributors))
	}

	// Verify contributor names
	names := make(map[string]bool)
	for _, c := range contributors {
		names[c.Name] = true
	}

	if !names["Alice"] || !names["Bob"] || !names["Charlie"] {
		t.Error("expected Alice, Bob, and Charlie in contributors")
	}

	// Verify Charlie has GitHub host detected
	for _, c := range contributors {
		if c.Name == "Charlie" {
			if c.Host != "github.com" {
				t.Errorf("Charlie's host = %q, want 'github.com'", c.Host)
			}
			if c.Username != "charlie" {
				t.Errorf("Charlie's username = %q, want 'charlie'", c.Username)
			}
		}
	}
}

func TestCommitInfo(t *testing.T) {
	commit := CommitInfo{
		Hash:        "abc123def456",
		ShortHash:   "abc123d",
		Subject:     "feat: add feature",
		Author:      "Test Author",
		AuthorEmail: "test@example.com",
	}

	if commit.Hash != "abc123def456" {
		t.Errorf("Hash = %q, want 'abc123def456'", commit.Hash)
	}
	if commit.ShortHash != "abc123d" {
		t.Errorf("ShortHash = %q, want 'abc123d'", commit.ShortHash)
	}
	if commit.Subject != "feat: add feature" {
		t.Errorf("Subject = %q, want 'feat: add feature'", commit.Subject)
	}
}

func TestBuildRemoteInfo(t *testing.T) {
	tests := []struct {
		name         string
		host         string
		owner        string
		repo         string
		wantProvider string
	}{
		{
			name:         "GitHub",
			host:         "github.com",
			owner:        "owner",
			repo:         "repo",
			wantProvider: "github",
		},
		{
			name:         "GitLab",
			host:         "gitlab.com",
			owner:        "group",
			repo:         "project",
			wantProvider: "gitlab",
		},
		{
			name:         "Codeberg",
			host:         "codeberg.org",
			owner:        "user",
			repo:         "repo",
			wantProvider: "codeberg",
		},
		{
			name:         "Bitbucket",
			host:         "bitbucket.org",
			owner:        "team",
			repo:         "repo",
			wantProvider: "bitbucket",
		},
		{
			name:         "Custom host",
			host:         "git.mycompany.com",
			owner:        "team",
			repo:         "project",
			wantProvider: "custom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildRemoteInfo(tt.host, tt.owner, tt.repo)

			if got.Provider != tt.wantProvider {
				t.Errorf("Provider = %q, want %q", got.Provider, tt.wantProvider)
			}
			if got.Host != tt.host {
				t.Errorf("Host = %q, want %q", got.Host, tt.host)
			}
			if got.Owner != tt.owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.owner)
			}
			if got.Repo != tt.repo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.repo)
			}
		})
	}
}

func TestKnownProviders(t *testing.T) {
	expected := map[string]string{
		"github.com":    "github",
		"gitlab.com":    "gitlab",
		"codeberg.org":  "codeberg",
		"gitea.io":      "gitea",
		"bitbucket.org": "bitbucket",
		"sr.ht":         "sourcehut",
	}

	for host, provider := range expected {
		if got := KnownProviders[host]; got != provider {
			t.Errorf("KnownProviders[%q] = %q, want %q", host, got, provider)
		}
	}
}

func TestGetCommitsWithMeta_MockSuccess(t *testing.T) {
	// Save and restore original function
	originalFn := GetCommitsWithMetaFn
	defer func() { GetCommitsWithMetaFn = originalFn }()

	// Mock the function
	GetCommitsWithMetaFn = func(since, until string) ([]CommitInfo, error) {
		return []CommitInfo{
			{Hash: "abc123", ShortHash: "abc123", Subject: "feat: test", Author: "Test", AuthorEmail: "test@example.com"},
			{Hash: "def456", ShortHash: "def456", Subject: "fix: bug", Author: "User", AuthorEmail: "user@example.com"},
		}, nil
	}

	commits, err := GetCommitsWithMetaFn("v1.0.0", "HEAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(commits) != 2 {
		t.Errorf("expected 2 commits, got %d", len(commits))
	}
}

func TestGetRemoteInfo_MockSuccess(t *testing.T) {
	// Save and restore original function
	originalFn := GetRemoteInfoFn
	defer func() { GetRemoteInfoFn = originalFn }()

	// Mock the function
	GetRemoteInfoFn = func() (*RemoteInfo, error) {
		return &RemoteInfo{
			Provider: "github",
			Host:     "github.com",
			Owner:    "testowner",
			Repo:     "testrepo",
		}, nil
	}

	remote, err := GetRemoteInfoFn()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remote.Owner != "testowner" {
		t.Errorf("Owner = %q, want 'testowner'", remote.Owner)
	}
}

func TestGetLatestTag_MockSuccess(t *testing.T) {
	// Save and restore original function
	originalFn := GetLatestTagFn
	defer func() { GetLatestTagFn = originalFn }()

	// Mock the function
	GetLatestTagFn = func() (string, error) {
		return "v1.0.0", nil
	}

	tag, err := GetLatestTagFn()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag != "v1.0.0" {
		t.Errorf("tag = %q, want 'v1.0.0'", tag)
	}
}

func TestGetContributors_MockSuccess(t *testing.T) {
	// Save and restore original function
	originalFn := GetContributorsFn
	defer func() { GetContributorsFn = originalFn }()

	// Mock the function
	GetContributorsFn = func(commits []CommitInfo) []Contributor {
		return []Contributor{
			{Name: "Test User", Username: "testuser", Email: "test@example.com", Host: "github.com"},
		}
	}

	commits := []CommitInfo{{Author: "Test", AuthorEmail: "test@example.com"}}
	contributors := GetContributorsFn(commits)
	if len(contributors) != 1 {
		t.Errorf("expected 1 contributor, got %d", len(contributors))
	}
}

func TestRemoteInfo_Fields(t *testing.T) {
	remote := RemoteInfo{
		Provider: "gitlab",
		Host:     "gitlab.example.com",
		Owner:    "mygroup",
		Repo:     "myproject",
	}

	if remote.Provider != "gitlab" {
		t.Errorf("Provider = %q, want 'gitlab'", remote.Provider)
	}
	if remote.Host != "gitlab.example.com" {
		t.Errorf("Host = %q, want 'gitlab.example.com'", remote.Host)
	}
	if remote.Owner != "mygroup" {
		t.Errorf("Owner = %q, want 'mygroup'", remote.Owner)
	}
	if remote.Repo != "myproject" {
		t.Errorf("Repo = %q, want 'myproject'", remote.Repo)
	}
}

func TestContributor_Fields(t *testing.T) {
	contrib := Contributor{
		Name:     "Alice Smith",
		Username: "alicesmith",
		Email:    "alice@example.com",
		Host:     "github.com",
	}

	if contrib.Name != "Alice Smith" {
		t.Errorf("Name = %q, want 'Alice Smith'", contrib.Name)
	}
	if contrib.Username != "alicesmith" {
		t.Errorf("Username = %q, want 'alicesmith'", contrib.Username)
	}
	if contrib.Email != "alice@example.com" {
		t.Errorf("Email = %q, want 'alice@example.com'", contrib.Email)
	}
	if contrib.Host != "github.com" {
		t.Errorf("Host = %q, want 'github.com'", contrib.Host)
	}
}
