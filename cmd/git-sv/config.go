package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strings"
	"sv4git/sv"

	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
)

// EnvConfig env vars for cli configuration
type EnvConfig struct {
	Home string `envconfig:"SV4GIT_HOME" default:""`

	MajorVersionTypes           []string          `envconfig:"MAJOR_VERSION_TYPES" default:""`
	MinorVersionTypes           []string          `envconfig:"MINOR_VERSION_TYPES" default:"feat"`
	PatchVersionTypes           []string          `envconfig:"PATCH_VERSION_TYPES" default:"build,ci,chore,docs,fix,perf,refactor,style,test"`
	IncludeUnknownTypeAsPatch   bool              `envconfig:"INCLUDE_UNKNOWN_TYPE_AS_PATCH" default:"true"`
	BreakingChangePrefixes      []string          `envconfig:"BRAKING_CHANGE_PREFIXES" default:"BREAKING CHANGE,BREAKING CHANGES"`
	IssueIDPrefixes             []string          `envconfig:"ISSUEID_PREFIXES" default:"jira,JIRA,Jira"`
	TagPattern                  string            `envconfig:"TAG_PATTERN" default:"%d.%d.%d"`
	ReleaseNotesTags            map[string]string `envconfig:"RELEASE_NOTES_TAGS" default:"fix:Bug Fixes,feat:Features"`
	ValidateMessageSkipBranches []string          `envconfig:"VALIDATE_MESSAGE_SKIP_BRANCHES" default:"master,develop"`
	CommitMessageTypes          []string          `envconfig:"COMMIT_MESSAGE_TYPES" default:"build,ci,chore,docs,feat,fix,perf,refactor,revert,style,test"`
	IssueKeyName                string            `envconfig:"ISSUE_KEY_NAME" default:"jira"`
	IssueRegex                  string            `envconfig:"ISSUE_REGEX" default:"[A-Z]+-[0-9]+"`
	BranchIssuePrefixRegex      string            `envconfig:"BRANCH_ISSUE_PREFIX_REGEX" default:"([a-z]+\\/)?"`
	BranchIssueSuffixRegex      string            `envconfig:"BRANCH_ISSUE_SUFFIX_REGEX" default:"(-.*)?"`
}

func loadEnvConfig() EnvConfig {
	var c EnvConfig
	err := envconfig.Process("SV4GIT", &c)
	if err != nil {
		log.Fatal(err.Error())
	}
	return c
}

// Config cli yaml config
type Config struct {
	Version       string                 `yaml:"version"`
	Versioning    sv.VersioningConfig    `yaml:"versioning"`
	Tag           sv.TagConfig           `yaml:"tag"`
	ReleaseNotes  sv.ReleaseNotesConfig  `yaml:"release-notes"`
	Branches      sv.BranchesConfig      `yaml:"branches"`
	CommitMessage sv.CommitMessageConfig `yaml:"commit-message"`
}

func getRepoPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.New(string(out))
	}
	return strings.TrimSpace(string(out)), nil
}

func loadConfig(filepath string) (Config, error) {
	content, rerr := ioutil.ReadFile(filepath)
	if rerr != nil {
		return Config{}, rerr
	}

	var cfg Config
	cerr := yaml.Unmarshal(content, &cfg)
	if cerr != nil {
		return Config{}, fmt.Errorf("could not parse config from path: %s, error: %v", filepath, cerr)
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Version: "1.0",
		Versioning: sv.VersioningConfig{
			UpdateMajor:   []string{},
			UpdateMinor:   []string{"feat"},
			UpdatePatch:   []string{"build", "ci", "chore", "docs", "fix", "perf", "refactor", "style", "test"},
			IgnoreUnknown: false,
		},
		Tag:          sv.TagConfig{Pattern: "%d.%d.%d"},
		ReleaseNotes: sv.ReleaseNotesConfig{Headers: map[string]string{"fix": "Bug Fixes", "feat": "Features", "breaking-change": "Breaking Changes"}},
		Branches: sv.BranchesConfig{
			PrefixRegex:  "([a-z]+\\/)?",
			SuffixRegex:  "(-.*)?",
			DisableIssue: false,
			Skip:         []string{"master", "main", "developer"},
		},
		CommitMessage: sv.CommitMessageConfig{
			Types: []string{"build", "ci", "chore", "docs", "feat", "fix", "perf", "refactor", "revert", "style", "test"},
			Scope: sv.CommitMessageScopeConfig{},
			Footer: map[string]sv.CommitMessageFooterConfig{
				"issue":           {Key: "jira", KeySynonyms: []string{"Jira", "JIRA"}},
				"breaking-change": {Key: "BREAKING CHANGE", KeySynonyms: []string{"BREAKING CHANGES"}},
			},
			Issue: sv.CommitMessageIssueConfig{Regex: "[A-Z]+-[0-9]+"},
		},
	}
}
