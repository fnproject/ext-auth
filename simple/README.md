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

# Login:
curl -H "Content-Type: application/json" -X POST -d '{"username":"johnny","password":"xyz"}' http://localhost:8080/v1/login

# Grab token returned from above and try to access endpoint again
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/v1/apps
# Success! :)
```

## For Development

```sh
cd main
SIMPLE_SECRET=ubersecret go run main.go
```
