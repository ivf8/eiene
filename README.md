# eiene

A simple shell in GO :candy:

## Building

### Requirements
- [GO]('https://go.dev') version >=1.23.3
- [cmake]('https://cmake.org/download') version >=3.30
- [make]('https://www.gnu.org/software/make/#download')


Clone this repository

```bash
git clone https://github.com/ivf8/eiene
```

Move into the `eiene` directory and create `build` directory.

```bash
cd eiene && mkdir build
```
Run the `cmake` command to create needed all build files.

```bash
cmake ..
```

Build the shell.

```bash
make build
```

Now you can run the shell

```bash
./eiene
```

## Testing

To run tests, just run `make test` in the directory with the build files.
`cmake` command must have been run in order to produce the build files.

```bash
cd build && make test
```

If you want verbose output for the tests, run `make test_verbose` and if you want to
see the test coverage run `make coverage`.

-----

:chipmunk: _*ivf8*_
