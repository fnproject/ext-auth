# A very simple (too simple) auth extension example

Can be used for testing or as an example to build more complete authentication.

## Usage

Simply add `github.com/treeder/fn-ext-auth/simple` to your ext.yaml file and build. [Learn more](https://github.com/fnproject/fn/blob/master/docs/operating/extending.md).

Required environment variables:

* SIMPLE_SECRET - a secret string to use for signing JWT tokens.

Try running the following commands to try it out:

```sh
# First, let's try to access and endpoint without credentials.
curl http://localhost:8080/v1/apps
# Fails... :(

# Login: this will create a user if one doesn't exist
curl -H "Content-Type: application/json" -X POST -d '{"username":"johnny","password":"xyz"}' http://localhost:8080/v1/login

# deploy a function
fn init --runtime go gofunc
cd gofunc
fn deploy --app myapp --local
# SHOULD FAIL

export FN_TOKEN=YOUR_TOKEN
fn deploy --app myapp --local
# SHOULD WORK

# Grab token returned from above and try to access another endpoint
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/v1/apps
# Success! :)

# now make another user
curl -H "Content-Type: application/json" -X POST -d '{"username":"tommy","password":"abc"}' http://localhost:8080/v1/login
export FN_TOKEN=TOMMY_TOKEN
fn deploy --app myapp --local
# SHOULD FAIL
fn deploy --app tommyapp --local
# SHOULD WORK
```

## For Development

```sh
cd main
SIMPLE_SECRET=ubersecret go run main.go
```
