# AWS SSO LOGIN

## Getting Started

This application depends on accounts being in your /home/{user}/.aws/config. This means you will need to run the following:

```
aws configure sso
```

You will also need to be logged into the sso. To do this run the following command:

```
aws sso login
```

Future development will provide a prompt to login if the current login does not exist.

Once you are logged in you can then run:

```
aws-sso-login
```

## Installation

This application is written in go. You can install the application by running the following command in the root of the directory:

```
go mod tidy
go build
```

This will produce the aws-sso-login binary which you can then move into a location that is in your PATH:

```
sudo mv aws-sso-login /usr/local/bin
```