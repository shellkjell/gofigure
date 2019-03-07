# gofigure

Package gofigure is a fully functional configuration file parser with bells on

## Getting Started

Want to get started with this simpler way of configuring software?
Good. You shall now write .fig files for all of your projects.

### Prerequisites

gofigure can be run 
* in a docker container - requires an installation of `docker`
* locally on your machine - requires an installation of `golang` (this implementation was written in version 1.10)

### Installing

```
git clone git@github.com:techbuddyab/gofigure
```

Now, you can choose to run the software locally if you installed golang. As such

```
# build with local go installation
make build-local
```

Upon which you will have a built binary in the project root folder.
The docker way has the same end result

```
# build in docker
make build
```

Now try your local copy of the software out with `make run` or `make run-local` depending on whether you chose the docker or local build way. Both should actually work after any build.

## Running the tests

Running the tests locally is easy. Just `go test` it!

## Built With

* [participle](https://github.com/alecthomas/participle) - The parser library used
* No other dependencies - and we aim to keep it that way.

## Contributing

No `CONTRIBUTING.md` yet, but feel free to submit an issue if you have problems or ideas. PRs are also welcome, please submit a short description of your work and we'll take it up together from there.

## Authors

* **Zakay Danial** - *Idea, initial structure and mastermind* - [Zakay](https://github.com/Zakay)
* **Simon Haak** - *Full implementation* - [shellkjell](https://github.com/shellkjell)

## License

This project is licensed under the Apache v2.0 License - see the [LICENSE.md](LICENSE.md) file for details

## Acknowledgments

* Examples from the alecthomas/participle repo were helpful - A truly dead simple parser.
