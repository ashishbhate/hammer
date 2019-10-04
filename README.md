# hammer
A library that allows querying of a large number of bitcoin address balances without
needing to set up a local Bitcoin node.

It works by querying a large number of bitcoin service APIs, automatically
distributing load and backing off from querying APIs to stay within their free
API rate limits.

This is a better Go implementation of the remote resources runner of the bcsearch (
[github](https://github.com/ashishbhate/bcsearch) /
[gitlab](https://gitlab.com/ashishbhate/bcsearch)
) library.

## Usage
The library needs no configuration. See the hammer cmd for example usage.

## Workers
See the repo's issues for upcoming support for APIs. The more workers there are, the higher the querying capacity of the library

## Issues

The repo issues are a fair representation of known issues and planned improvements. The [GitLab](https://gitlab.com/ashishbhate/hammer) repo is the canonical repository. The [GitHub](https://github.com/ashishbhate/hammer) repo is a mirror.
