# FTL for Visual Studio Code

[The VS Code FTL extension](https://marketplace.visualstudio.com/items?itemName=FTL.ftl)
provides support for
[FTL](https://github.com/block/ftl) within VSCode including LSP integration and useful commands for manageing FTL projects.

## Requirements

- FTL 0.169.0 or newer

## Quick Start

1.  Install [FTL](https://github.com/block/ftl) 0.169.0 or newer.

2.  Install this extension.

3.  Open any FTL project with a `ftl-project.toml` or `ftl.toml` file.

## Settings

Configure the FTL extension by setting the following options in your Visual Studio Code settings.json:

- `ftl.executablePath`: Specifies the path to the FTL executable. The default is "ftl", which uses the system's PATH to find the executable.

  > [!IMPORTANT]
  > If you have installed FTL with hermit (or other dependency management tools), you may need to specify the path to the FTL binary in the extension settings.

  ```json
  {
    "ftl.executablePath": "bin/ftl"`
  }
  ```

- `ftl.lspCommandFlags`: Defines flags to pass to the FTL executable when starting the development environment.

  ```json
  {
    "ftl.lspCommandFlags": []
  }
  ```
