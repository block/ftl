<img height="80px" align="right" src="https://www.wtamu.edu/~cbaird/sq/images/fasterthanlight.png" alt="Public Domain Image, source: Christopher S. Baird"/>

<br />

# FTL [![CI](https://github.com/block/ftl/actions/workflows/ci.yml/badge.svg)](https://github.com/block/ftl/actions/workflows/ci.yml) [![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

## Documentation

https://block.github.io/ftl/

## Getting started

### Install ftl, for example on macos:

```sh
brew tap block/ftl && brew install ftl
```

### Create a sample project (Go)

<!-- This is for [mdcode](https://github.com/szkiba/mdcode) to test snippets in our README. -->

<!--<script type="text/markdown">
```sh file=test.sh outline=true
#!/bin/bash
set -Eeuxo pipefail

just build ftl
export PATH="$(git rev-parse --show-toplevel)/build/release:$PATH"

pwd

# #region init
# #endregion

(
# #region start
# #endregion
) &
pid="$!"
trap "kill $pid" EXIT ERR INT

diff -u <(
(
# #region call
# #endregion
) | tee /dev/stderr
) <(echo '{"message":"Hello, Bob!"}')
```
</script>-->

```sh file=test.sh region=init
ftl init myproject
cd myproject
ftl module new go alice
```

### Build and deploy the module

Start FTL:

```sh file=test.sh region=start
ftl dev --wait-for=alice
```

Now let's call a verb:

```sh file=test.sh region=call
ftl call alice.hello '{name: "Bob"}'
```

## Project Resources

| Resource                                   | Description                                                                   |
| ------------------------------------------ | ----------------------------------------------------------------------------- |
| [CODEOWNERS](./CODEOWNERS)                 | Outlines the project lead(s)                                                  |
| [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) | Expected behavior for project contributors, promoting a welcoming environment |
| [CONTRIBUTING.md](./CONTRIBUTING.md)       | Developer guide to build, test, run, access CI, chat, discuss, file issues    |
| [GOVERNANCE.md](./GOVERNANCE.md)           | Project governance                                                            |
| [LICENSE](./LICENSE)                       | Apache License, Version 2.0                                                   |
