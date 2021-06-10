## Installation

Add Lefthook to your system or build it from sources.

### go

```bash
go get github.com/evilmartians/lefthook
```

### npm or yarn

```bash
npm i @arkweid/lefthook --save-dev
# or yarn:
yarn add -D @arkweid/lefthook
```

NOTE: if you install it this way you should call it with `npx` or `yarn` for all listed examples below. (for example: `lefthook install` -> `npx lefthook install`)

You can also install lefthook as a global dependency

```bash
npm install -g @arkweid/lefthook
```


### Rubygems

```bash
gem install lefthook
```

### Homebrew for MacOS and Linux

```bash
brew install lefthook
```

### AUR for Arch

You can install lefthook [package](https://aur.archlinux.org/packages/lefthook) from AUR

### Snap for Linux

```sh
snap install --classic lefthook
```

### APT packages for Debian/Ubuntu Linux

```sh
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.deb.sh' | sudo -E bash
sudo apt install lefthook
```

See all instructions: https://cloudsmith.io/~evilmartians/repos/lefthook/setup/#formats-deb

[![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com "Debian package repository hosting is graciously provided by Cloudsmith")

### RPM packages for CentOS/Fedora Linux

```sh
curl -1sLf 'https://dl.cloudsmith.io/public/evilmartians/lefthook/setup.rpm.sh' | sudo -E bash
sudo yum install lefthook
```

See all instructions: https://cloudsmith.io/~evilmartians/repos/lefthook/setup/#repository-setup-yum

[![Hosted By: Cloudsmith](https://img.shields.io/badge/OSS%20hosting%20by-cloudsmith-blue?logo=cloudsmith&style=flat-square)](https://cloudsmith.com "RPM package repository hosting is graciously provided by Cloudsmith")

### Anything else

Or take it from [binaries](https://github.com/evilmartians/lefthook/releases) and install manually

## Scenarios

### Examples

We have a directory with few examples. You can check it [here](https://github.com/evilmartians/lefthook/tree/master/examples).

### First time user

Initialize lefthook with the following command

```bash
lefthook install
```

It creates `lefthook.yml` in the project root directory

Register your hook (You can choose any hook from [this list](https://git-scm.com/docs/githooks))
In our example it `pre-push` githook:

```bash
lefthook add pre-push
```

Describe pre-push commands in `lefthook.yml`:

```yml
pre-push: # githook name
  commands: # list of commands
    packages-audit: # command name
      run: yarn audit # command for execution
```

That's all! Now on `git push` the `yarn audit` command will be run.
If it fails the `git push` will be interrupted.

### If you already have a lefthook config file

Just initialize lefthook to make it work :)

```bash
lefthook install
```

## More options

## Use glob patterns to choose what files you want to check

```yml
# lefthook.yml

pre-commit:
  commands:
    lint:
      glob: "*.{js,ts,jsx,tsx}"
      run: yarn eslint
```

## Select specific file groups

In some cases you want to run checks only against some specific file group.
For example: you may want to run eslint for staged files only.

There are two shorthands for such situations:
`{staged_files}` - staged git files which you try to commit

`{all_files}` - all tracked files by git

```yml
# lefthook.yml

pre-commit:
  commands:
    frontend-linter:
      glob: "*.{js,ts,jsx,tsx}" # glob filter for list of files
      run: yarn eslint {staged_files} # {staged_files} - list of files
    backend-linter:
      glob: "*.rb" # glob filter for list of files
      exclude: "application.rb|routes.rb" # regexp filter for list of files
      run: bundle exec rubocop --force-exclusion {all_files} # {all_files} - list of files
```

Note: If using `all_files` with RuboCop, it will ignore RuboCop's `Exclude` configuration setting. To avoid this, pass `--force-exclusion`.

## Custom file list

Lefthook can be even more specific in selecting files.
If you want to choose diff of all changed files between the current branch and master branch you can do it this way:

```yml
# lefthook.yml

pre-push:
  commands:
    frontend-style:
      files: git diff --name-only master # custom list of files
      glob: "*.js"
      run: yarn stylelint {files}
```

`{files}` - shorthand for a custom list of files

## Git hook argument shorthands in commands

If you want to use the original Git hook arguments in a command you can do it
using the indexed shorthands:

```yml
# lefthook.yml

# Note: commit-msg hook takes a single parameter,
# the name of the file that holds the proposed commit log message.
# Source: https://git-scm.com/docs/githooks#_commit_msg
commit-msg:
  commands:
    multiple-sign-off:
      run: 'test $(grep -c "^Signed-off-by: " {1}) -lt 2'
```
`{0}` - shorthand for the single space-joint string of Git hook arguments

`{i}` - shorthand for the i-th Git hook argument

## Managing scripts

If you run `lefthook add` command with `-d` flag, lefthook will create two directories where you can put scripts and reference them from `lefthook.yml` file.

Example:
Let's create `commit-msg` hook with `-d` flag

```bash
lefthook add -d commit-msg
```

This command will create `.lefthook/commit-msg` and `.lefthook-local/commit-msg` dirs.

The first one is for common project level scripts.
The second one is for personal scripts. It would be a good idea to add dir`.lefthook-local` to `.gitignore`.

Create scripts `.lefthook/commit-msg/hello.js` and `.lefthook/commit-msg/hi.rb`

```yml
# lefthook.yml

commit-msg:
  scripts:
    "hello.js":
      runner: node
    "hi.rb":
      runner: ruby
```

### Bash script example

Let's create a bash script to check commit templates `.lefthook/commit-msg/template_checker`:

```bash
INPUT_FILE=$1
START_LINE=`head -n1 $INPUT_FILE`
PATTERN="^(TICKET)-[[:digit:]]+: "
if ! [[ "$START_LINE" =~ $PATTERN ]]; then
  echo "Bad commit message, see example: TICKET-123: some text"
  exit 1
fi
```

Now we can ask lefthook to run our bash script by adding this code to
`lefthook.yml` file:

```yml
# lefthook.yml

commit-msg:
  scripts:
    "template_checker":
      runner: bash
```

When you try to commit `git commit -m "haha bad commit text"` script `template_checker` will be executed. Since commit text doesn't match the described pattern the commit process will be interrupted.

## Bash script example with Commitlint

Let's create a bash script to check conventional commit status `.lefthook/commit-msg/commitlint.sh`:

```bash
echo $(head -n1 $1) | npx commitlint --color
```

Now we can ask lefthook to run our bash script by adding this code to
`lefthook.yml` file:

```yml
# lefthook.yml

commit-msg:
  scripts:
    "commitlint.sh":
      runner: bash
```

When you try to commit `git commit -m "haha bad commit text"` script `commitlint.sh` will be executed. Since commit text doesn't match the default config or custom config that you setup for `commitlint`, the process will be interrupted.

## Local config

We can use `lefthook-local.yml` as local config. Options in this file will overwrite options in `lefthook.yml`. (Don't forget to add this file to `.gitignore`)

## Skipping commands

You can skip commands by `skip` option:

```yml
# lefthook-local.yml

pre-push:
  commands:
    packages-audit:
      skip: true
```

## Skipping commands during rebase or merge

You can skip commands during rebase and/or merge by the same `skip` option:

```yml
pre-push:
  commands:
    packages-audit:
      skip: merge

# or

pre-push:
  commands:
    packages-audit:
      skip:
        - merge
        - rebase
```

## Skipping commands by tags

If we have a lot of commands and scripts we can tag them and run skip commands with a specific tag.

For example, if we have `lefthook.yml` like this:

```yml
# lefthook.yml

pre-push:
  commands:
    packages-audit:
      tags: frontend security
      run: yarn audit
    gems-audit:
      tags: backend security
      run: bundle audit
```

You can skip commands by tags:

```yml
# lefthook-local.yml

pre-push:
  exclude_tags:
    - frontend
```

## Piped option
If any command in the sequence fails, the other will not be executed.
```yml
# lefthook.yml

database:
  piped: true
  commands:
    1_create:
      run: rake db:create
    2_migrate:
      run: rake db:migrate
    3_seed:
      run: rake db:seed
```

## Extends option
If you need to extend config from some another place, just add top level:
```yml
# lefthook.yml

extends:
  - $HOME/work/lefthook-extend.yml
  - $HOME/work/lefthook-extend-2.yml
```
NOTE: Files for extend should have name NOT a "lefthook.yml" and should be unique.

## Referencing commands from lefthook.yml

If you have the following config

```yml
# lefthook.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: node
```

You can wrap it in docker runner locally:

```yml
# lefthook-local.yml

pre-commit:
  scripts:
    "good_job.js":
      runner: docker run -it --rm <container_id_or_name> {cmd}
```

`{cmd}` - shorthand for the command from `lefthook.yml`

## Run githook group directly

```bash
lefthook run pre-commit
```

## Parallel execution

You can enable parallel execution if you want to speed up your checks.
Lets get example from [discourse](https://github.com/discourse/discourse/blob/master/.travis.yml#L77-L83) project.

```
bundle exec rubocop --parallel && \
bundle exec danger && \
yarn eslint --ext .es6 app/assets/javascripts && \
yarn eslint --ext .es6 test/javascripts && \
yarn eslint --ext .es6 plugins/**/assets/javascripts && \
yarn eslint --ext .es6 plugins/**/test/javascripts && \
yarn eslint app/assets/javascripts test/javascripts
```

Rewrite it in lefthook custom group. We call it `lint`:

```yml
# lefthook.yml

lint:
  parallel: true
  commands:
    rubocop:
      run: bundle exec rubocop --parallel
    danger:
      run: bundle exec danger
    eslint-assets:
      run: yarn eslint --ext .es6 app/assets/javascripts
    eslint-test:
      run: yarn eslint --ext .es6 test/javascripts
    eslint-plugins-assets:
      run: yarn eslint --ext .es6 plugins/**/assets/javascripts
    eslint-plugins-test:
      run: yarn eslint --ext .es6 plugins/**/test/javascripts
    eslint-assets-tests:
      run: yarn eslint app/assets/javascripts test/javascripts
```

Then call this group directly:

```
lefthook run lint
```

## Complete example

```yml
# lefthook.yml
color: false
extends: $HOME/work/lefthook-extend.yml

pre-commit:
  commands:
    eslint:
      glob: "*.{js,ts,jsx,tsx}"
      run: yarn eslint {staged_files}
    rubocop:
      tags: backend style
      glob: "*.rb"
      exclude: "application.rb|routes.rb"
      run: bundle exec rubocop --force-exclusion {all_files}
    govet:
      tags: backend style
      files: git ls-files -m
      glob: "*.go"
      run: go vet {files}

  scripts:
    "hello.js":
      runner: node
    "any.go":
      runner: go run

  parallel: true
```

```yml
# lefthook-local.yml

pre-commit:
  exclude_tags:
    - backend

  scripts:
    "hello.js":
      runner: docker run -it --rm <container_id_or_name> {cmd}
  commands:
    govet:
      skip: true
```

## Skip lefthook execution

We can set env variable `LEFTHOOK` to zero for that

```bash
LEFTHOOK=0 git commit -am "Lefthook skipped"
```

## Skip some tags on the fly

Use LEFTHOOK_EXCLUDE={list of tags to be excluded} for that

```bash
LEFTHOOK_EXCLUDE=ruby,security git commit -am "Skip some tag checks"
```

## Concurrent files overrides

To prevent concurrent problems with read/write files try `flock`
utility.

```yml
# lefthook.yml

graphql-schema:
  glob: "{Gemfile.lock,app/graphql/**/*}"
  run: flock webpack/application/typings/graphql-schema.json yarn typings:update && git diff --exit-code --stat HEAD webpack/application/typings
frontend-tests:
  glob: "**/*.js"
  run: flock -s webpack/application/typings/graphql-schema.json yarn test --findRelatedTests {files}
frontend-typings:
  glob: "**/*.js"
  run: flock -s webpack/application/typings/graphql-schema.json yarn run flow focus-check {files}
```

## Capture ARGS from git in the script

Example script for `prepare-commit-msg` hook:

```bash
COMMIT_MSG_FILE=$1
COMMIT_SOURCE=$2
SHA1=$3

# ...
```

## Change directory for script files

You can do this through this config keys:

```yml
# lefthook.yml

source_dir: ".lefthook"
source_dir_local: ".lefthook-local"
```

## Manage verbosity

You can manage the verbosity using the `skip_output` config.

Possible values are `meta,success,failure,summary`.

This config quiets all outputs except failures:

```yml
# lefthook.yml

skip_output:
  - meta
  - success
  - summary
```

You can also do this with an environment variable:
```bash
export LEFTHOOK_QUIET="meta,success,summary"
```

## CI integration

Enable `CI` env variable if it doens't exists on your service by default.

## Disable colors

By agrs:
```bash
lefthook --no-colors run pre-commit
```
By config `lefthook.yml`, just add the option:
```yml
colors: false
```

## Version

```bash
lefthook version
```

## Uninstall

```bash
lefthook uninstall
```

## More info
Have a question? Check the [wiki](https://github.com/evilmartians/lefthook/wiki).
