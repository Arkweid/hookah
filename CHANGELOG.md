# Change log

## master (unreleased)

- [PR](https://github.com/evilmartians/lefthook/pull/185) Add `.lefthookrc` to fix incorrect PATH @skryukov
- [PR](https://github.com/evilmartians/lefthook/pull/182) Support git workspaces and submodules @skryukov


# 0.7.4 (2021-04-30)

- [PR](https://github.com/evilmartians/lefthook/pull/171) Improve check for installed git @DmitryTsepelev
- [PR](https://github.com/evilmartians/lefthook/pull/169) Create .git/hooks directory when it does not exist @DmitryTsepelev

# 0.7.3 (2021-04-23)

- [PR](https://github.com/evilmartians/lefthook/pull/168) Package versions for all architectures (x86_64, ARM64, x86) into Ruby gem and NPM package @Envek
- [PR](https://github.com/evilmartians/lefthook/pull/167) Fix golang 15+ build @skryukov

# 0.7.2 (2020-02-02)

- [PR](https://github.com/evilmartians/lefthook/pull/126) Feature multiple extends. Thanks @Evilweed

- [PR](https://github.com/evilmartians/lefthook/pull/124) Fix `npx` when only `yarn` exists. Thanks @dotterian

- [PR](https://github.com/evilmartians/lefthook/pull/116) Fix use '-h' for robust lefthook. Thanks @fahrinh

# 0.7.1 (2020-02-02)

- [PR](https://github.com/evilmartians/lefthook/pull/108) Fix `sh` dependency on windows when executing `git`. Thanks @lionskape

- [PR](https://github.com/evilmartians/lefthook/pull/103) Add possibility for using `yaml` and `yml` extension for config. Thanks @rbUUbr

# 0.7.0 (2019-12-14)

- [PR](https://github.com/evilmartians/lefthook/pull/98) Support relative roots for monorepos. Thanks @jsmestad

# 0.6.7 (2019-12-14)

- [Commit](https://github.com/evilmartians/lefthook/commit/e898b5c8ba56c4d6f29a4d1f433baa1779a0845b)
Skip before executing command

- [PR](https://github.com/evilmartians/lefthook/pull/94) Add option --keep-config. Thanks @justinasposiunas

- [Commit](https://github.com/evilmartians/lefthook/commit/d79a3a46e7d1ee709b97e97f823bfd27e9466eff)
Check if shell is non interactive

# 0.6.6 (2019-12-03)

- [PR](https://github.com/evilmartians/lefthook/pull/94) Use eval instead of exec; Enable tty for the shell. Thanks @ssnickolay

# 0.6.5 (2019-11-15)

- [PR](https://github.com/evilmartians/lefthook/pull/89) Add support for git-worktree. Thanks @f440

- [Commit](https://github.com/evilmartians/lefthook/commit/48702a0806d2b2eab13636ba56b0e0b99f346f1c)
Commands and Scripts now can catch Stdin

- [Commit](https://github.com/evilmartians/lefthook/commit/9a226842292ff1dda0f2273b66a0799988aa5289)
Add partial support for monorepos and command execution not from project root

# 0.6.4 (2019-11-08)

- [PR](https://github.com/evilmartians/lefthook/pull/84) Fix return value from shell exit. Thanks @HaiD84

- [PR](https://github.com/evilmartians/lefthook/pull/86) Support postinstall script for npm installation for monorepos. Thanks @sHooKDT

- [PR](https://github.com/evilmartians/lefthook/pull/82) Now relative path to scripts supported. Thanks @AlexeyMatskevich

- [Commit](https://github.com/evilmartians/lefthook/pull/80/commits/1a4b0ee155eb66ae6f3c365164012bee9332605a)
Option `extends` for top level config added. Now you can merge some settings from different places:
```yml
extends: $HOME/work/lefthook-extend.yml
```

- [Commit](https://github.com/evilmartians/lefthook/commit/83cf818106dbf222ea33ba86aafce8f30d7cb5a9)
Add examples to generated lefthook.yml

## 0.6.3 (2019-07-15)

- [Commit](https://github.com/evilmartians/lefthook/commit/0426936f48f248221126f15619932b0dc8c54d7a) Add `-a` means `aggressive` strategy for `install` command
```bash
lefthook install -a # clear .git/hooks dir and reinstall lefthook hooks
```

- [Commit](https://github.com/evilmartians/lefthook/commit/5efb0677a4a9ec1728d3cf1a083075e23315a796) Add Lefthook version indicator for commands and script execution

- [Commit](https://github.com/evilmartians/lefthook/commit/8b55d91eed46643a1674bd4ad96fa211a177e159) Remove `npx` as dependency from node wrapper

Now we will call directly binary from `./node_modules`

- [Commit](https://github.com/evilmartians/lefthook/commit/76ffed4c698bc074984e91f5610c0b98784bd10b) Add `-f` means `force` strategy for `install` command

```bash
lefthook install -f # reinstall lefthook hooks without sync info check
```

- PR [#27](https://github.com/evilmartians/lefthook/pull/27) Move LEFTHOOK env check in hooks files

Now if LEFTHOOK=0 we will not call the binary file

- PR [#26](https://github.com/evilmartians/lefthook/pull/26) + [commit](https://github.com/evilmartians/lefthook/commit/afd67f94631a10975209ed4c5fabc763f44280eb) Add `{push_files}` shortcut

Add shortcut `{push_files}`

```
pre-commit:
  commands:
    rubocop:
      run: rubocop {push_files}
```
It same as:
```
pre-commit:
  commands:
    rubocop:
      files: git diff --name-only HEAD @{push} || git diff --name-only HEAD master
      run: rubocop {push_files}
```

- [Commit](https://github.com/evilmartians/lefthook/commit/af087b032a14952aa1dd235a3d0b5a51bc760a10) Add `min_version` option

You can mark your config for minimum Lefthook version:
```
min_version: 0.6.1
```

## 0.6.0 (2019-07-10)

- PR [#24](https://github.com/palkan/logidze/pull/110) Wrap `run` command in shell context.

Now in `run` option available `sh` syntax.

```
pre-commit:
  commands:
    bashed:
      run: rubocop -a && git add
```
Will be executed in this way:
```
sh -c "rubocop -a && git add"
```

- PR [#23](https://github.com/evilmartians/lefthook/pull/24) Search Lefthook in Gemfile.

Now it's possible to use Lefthook from Gemfile.

```ruby
# Gemfile

gem 'lefthook'
```
