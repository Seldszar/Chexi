# Chexi

> Microservice providing presence information about a Roblox user

## Requirements

- [Go](https://go.dev/dl)

## Usage

```sh
$ go run main.go --user=<USER_ID>
```

Presence information might be unavailable sometimes, so you will have to provide a security token from an active `.ROBLOSECURITY` cookie in order to fix issue.

```sh
$ go run main.go --user=<USER_ID> --token=<TOKEN>
```

## License

Copyright (c) 2023-present Alexandre Breteau

This software is released under the terms of the MIT License.
See the [LICENSE](LICENSE) file for further information.
