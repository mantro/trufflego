Search for secrets in a directory

## Installation

```bash
brew tap mantro/trufflego
brew install trufflego
```

## Developer setup

```console
foo@bar:~$ brew install golang
foo@bar:~$ ./install.sh
```

## Help

```console
foo@bar:~$ trufflego -h
Usage:
  trufflego [OPTIONS] [directory]

Application Options:
  -t, --threshold= Default 4.8 (higher->lesser detections) (default: 4.8)
  -m, --minimum=   Minimum length of detected passwords (default: 12)

Help Options:
  -h, --help       Show this help message

Arguments:
  directory:       Path to start searching in
```
