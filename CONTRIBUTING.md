# Contributing

## Development guidelines

There are many great development guidelines available.

We refer to:

- https://github.com/hyperledger/fabric-x-committer/blob/main/guidelines.md

## Update dependencies

### Gomate

Our repository contains a multi go modules project. To simplify our life when dealing with go dependencies, we provide a little helper named `gomate.sh` (pronounced: goo - maa - te, like german Tomate).

#### Example usage

Update a specific dependency with a given version XXXX.

```bash
./gomate.sh update github.com/hyperledger-labs/fabric-smart-client@XXXX
./gomate.sh tidy
```

Update a specific dependency to the latest available version;

```bash
./gomate.sh update github.com/hyperledger-labs/fabric-smart-client
./gomate.sh tidy
```

Update a specific dependency to the latest available version;

```bash
./gomate.sh update github.com/hyperledger-labs/fabric-smart-client
./gomate.sh tidy
```

Update all dependencies - very brutal - not recommended!

```bash
./gomate.sh update
./gomate.sh tidy
```

Just run go mod tidy everywhere

```bash
./gomate.sh tidy
```

You can also ask for help

```bash
./gomate.sh help
```
