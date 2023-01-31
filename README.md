# mergemate
A TUI (text-based user interface) app that helps manage your merge requests on GitLab.

# Building
1. [Intall go 1.19](https://go.dev/doc/install) 
2. checkout repository
3. Run `go build -o mergeman cmd/merge/main.go` 

# Configuration
mergemate can be configured through configuration file and environment variables. Both approaches can be mixed together.

Configuration file should be placed under given location: `$XDG_CONFIG_HOME/mergemate/mergemate_config.env`.
`$XDG_CONFIG_HOME` is platform dependent, here are exact locations for mergemate's configuration file under various oparting systems:

| OS                | Config file path                                               |
|-------------------|----------------------------------------------------------------|
| Unix              | `~/.config/mergemate/mergemate_config.env`                     |
| macOS             | `~/Library/Application Support/mergemate/mergemate_config.env` |
| Microsoft Windows | `%LOCALAPPDATA%\mergemate\mergemate_config.env`                |

Configuration file should consist of key/value pairs separated by equal signs.

Following configuration options are supported:

| Option                               | Required | Default value | Description                                                                                                               |
|--------------------------------------|----------|:--------------|---------------------------------------------------------------------------------------------------------------------------|
| MERGEMATE_GITLAB_URL                 | YES      | -             | Your gitlab instance URL.                                                                                                 |
| MERGEMATE_API_TOKEN                  | YES      | -             | Your gitlab api token: https://docs.gitlab.com/ee/user/profile/personal_access_tokens.html#create-a-personal-access-token |
| MERGEMATE_USER_NAME                  | YES      | -             | Your gitlab user name.                                                                                                    |
| MERGEMATE_PROJECT_NAME               | YES      | -             | Name of the project where merge requests will be managed                                                                  |
| MERGEMATE_SLB_BRANCH_PREFIX          | YES      | -             | Branch prefix you use to distinguish your branches from those of your teammates.                                          |
| MERGEMATE_MERGE_JOB_INTERVAL_SECONDS | NO       | 60            | Time between two executions of background merge job.                                                                      |
| MERGEMATE_TARGET_BRANCH_PREFIXES     | NO       | ""            | Comma separated list of prefixes that match branches which should be shown on target branch list, i.e, master,Version_.   |

Empty configuration file template:
```
MERGEMATE_GITLAB_URL=
MERGEMATE_API_TOKEN=
MERGEMATE_USER_NAME=
MERGEMATE_PROJECT_NAME=
MERGEMATE_SLB_BRANCH_PREFIX=
```
# Troubleshooting
All performed actions and errors are written into a logfile. In case of errors the logfile should be used to investigate turn of events.  

Logfile is stored under given location: `$XDG_STATE_HOME/mergemate/debug.log`.
`$XDG_STATE_HOME` is platform dependent, here are exact locations for mergemate's logfile under various oparting systems:

| OS                | Config file path                                |
|-------------------|-------------------------------------------------|
| Unix              | `~/.local/state/mergemate/debug.log`            |
| macOS             | `~/Library/Application Support/mergemate/debug` |
| Microsoft Windows | `%LOCALAPPDATA%\mergemate\debug.log`            |




