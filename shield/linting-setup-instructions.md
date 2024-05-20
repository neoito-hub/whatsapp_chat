# Linting Setup

Pre commit Setup - https://github.com/go-courier/husky
Linting - https://golangci-lint.run/

## Pre commit Setup

https://pre-commit.com/

For mac
```
brew install pre-commit
```

For others

```
pip install pre-commit
```

Run

pre-commit install


## Linting Setup

### linux and windows

curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.45.2

golangci-lint --version

### In mac

```
brew install golangci-lint
brew upgrade golangci-lint
```

Inside CI runners
```
# binary will be $(go env GOPATH)/bin/golangci-lint
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.45.2

# or install it into ./bin/
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.45.2

# In alpine linux (as it does not come with curl by default)
wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s v1.45.2

golangci-lint --version
```