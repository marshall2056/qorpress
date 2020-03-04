# QorPress - go powered blog engine

## Desscription
Simply, a go wordpress clone with amazeballs inside

## Features
- Admin interface
- Dynamic front-end
- Smart SEO engine
- ...

## Screenshots
[to do]

## Deploy

In this section, we will explain how to setup an instance of QorPress...

### Pre-requisites

You can run QorPress in 2 modes, inside a docker container or locally. Running from docker allows to remove the manual installation of all sub-requirements listed here, as the local deploy is quite convenient for developpement purposes.

#### Running from docker
* Docker
* Docker-Compose

#### Running locally
* Go 1.13
* Mysql/PostgreSQL/Sqlite3
* Manticoresearch 

### Install
```bash
cd $GOPATH/src
mkdir -p github.com/x0rzkov
git clone --recursive --depth=1 https://github.com/x0rzkov/qorpress
cd qorpress
GO111MODULE=off go get -u -f github.com/qorpress/bindatafs/...
bindatafs config/bindatafs


go mod tidy -v
go mod vendor
go install ./cmd/...
```

### Run
```bash
go run -mod vendor *.go
```

#### Generate Fake data
```bash
go run cmd/tools/qp-kitploit/main.go
```

## Todos
* ~~add an autocert library/tool for letsencrypt~~
* create a plugin registry able for custom models/controllers

## Bugs
* ~~to do~~
* to do

## Contribute
Please Contribute by creating a fork of this repository.  
Follow the instructions here: https://help.github.com/articles/fork-a-repo/
