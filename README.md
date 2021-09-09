# gh-terasology - GitHub CLI extension for Terasology

This is a proof-of-concept [extension](https://github.blog/2021-08-24-github-cli-2-0-includes-extensions/) for `gh`, the GitHub CLI tool.
The concept of those extensions is quite simple, `gh` basically acts as plug-in manager and allows to call an exectuable as extension. 
See [Creating GitHub CLI Extensions](https://docs.github.com/en/github-cli/github-cli/creating-github-cli-extensions) for more details.

The **`terasology` extension** is a simple Bash script. It tries to detect the Terasology workspace root directory (by looking for the respective `settings.gradle` file).

## Getting Started

You can find more information on how to install and use extensions in general under [Using GitHub CLI Extensions](https://docs.github.com/en/github-cli/github-cli/using-github-cli-extensions).
To test this extension, you can easily install it via 

```
gh extension install skaldarnar/gh-terasology
```

## Usage

As this extensions is a proof-of-concept for testing out the capabilities of `gh` extensions for the use with Terasology there's not much to find here yet.

You can initialize the modules in the local workspace to a distribution from the [Index](https://github.com/Terasology/Index) repo by passing in a single positional argument:

```
gh terasology <distro>
```

For instance, to clone the smaller Iota distro run

```
gh terasology iota
```
