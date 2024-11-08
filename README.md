# Kubernetes Context Switcher

A simple Go application to switch between Kubernetes contexts interactively using `fzf`.

## Features

- List all available Kubernetes contexts.
- Filter contexts based on a search term.
- Select a context interactively using `fzf`.
- Option to unset the current context.

## Prerequisites

- Go (1.16 or later) environment setup
- kubectl
- fzf

## Installation

```bash
go install github.com/rymdlego/kubernetes-context-switcher@latest
```

Optional alias:

```bash
alias kx=kubernetes-context-switcher
```

## Usage

Run the application with an optional search term:
(assuming we have the alias setup)

```sh
kx [search-term]
```
