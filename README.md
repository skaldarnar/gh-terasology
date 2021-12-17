# gh-terasology - GitHub CLI extension for Terasology

This is a proof-of-concept [extension](https://github.blog/2021-08-24-github-cli-2-0-includes-extensions/) for `gh`, the GitHub CLI tool.
The concept of those extensions is quite simple, `gh` basically acts as plug-in manager and allows to call an executable as extension.
See [Creating GitHub CLI Extensions](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) for more details.

The **`terasology` extension** is a [precompiled Go extension](https://github.com/cli/gh-extension-precompile).

## Getting Started

You can find more information on how to install and use extensions in general under [Using GitHub CLI Extensions](https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions).
To test this extension, you can easily install it via

```
gh extension install skaldarnar/gh-terasology
```

## Usage

As this extensions is a proof-of-concept for testing out the capabilities of `gh` extensions for the use with Terasology there's not much to find here yet.

To get the changelog for the current repository since a specific date use the `--since` parameter:

```
gh terasology changelog --since 2021-12-01
```

To select a different single repository use the `--repo` (or `-R`) flag:

```
gh terasology changelog --repo MovingBlocks/TerasologyLauncher
```

To compile a changelog for all repositories of an organization run the following.
This requires an explicit `--since` flag as the "latest release" cannot be derived in this case.

```
gh terasology changelog --repo Terasology --since 2021-12-01
```

## Contributing

This extension is being build on [Go](https://go.dev/).
For local development make sure that you [download and install Go](https://go.dev/doc/install) for your platform.

For any changes to take effect the extension has to be compiled into a self-contained executable.
This is done by simply running :

```
go build
```

You can install the extension locally from source by running

```
gh extension install .
```

It will automatically update whenever the executable is rebuild.
No need to re-install over and over again.

## Roadmap

- assemble ready-made _changelogs_ for releases of Terasology
- _multi-repo management_, e.g., to update common configurations (settings, topics, ...)
- _release management_, e.g., automatically perform actions required for game or module releases
- _workspace management_, e.g., pinning of repository state, checkout by date, ...
