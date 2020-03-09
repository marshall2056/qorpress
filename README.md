# QorPress 

<p align="center">
    <a href="https://travis-ci.com/qorpress/qorpress"><img src="https://img.shields.io/travis/com/qorpress/qorpress.svg" /></a>
    <a href="https://cloud.drone.io/qorpress/qorpress"><img src="https://cloud.drone.io/api/badges/qorpress/qorpress/status.svg?ref=refs/heads/master" /></a>
</p>

<p align="center">
    <a href="https://github.com/qorpress/qorpress" alt="github all releases"><img src="https://img.shields.io/github/downloads/qorpress/qorpress/total.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github latest release"><img src="https://img.shields.io/github/downloads/qorpress/qorpress/latest/total.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github tag"><img src="https://img.shields.io/github/tag/qorpress/qorpress.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github release"><img src="https://img.shields.io/github/release/qorpress/qorpress.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github pre release"><img src="https://img.shields.io/github/release/qorpress/qorpress/all.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github fork"><img src="https://img.shields.io/github/forks/qorpress/qorpress.svg?style=social&label=Fork" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github stars"><img src="https://img.shields.io/github/stars/qorpress/qorpress.svg?style=social&label=Star" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github watchers"><img src="https://img.shields.io/github/watchers/qorpress/qorpress.svg?style=social&label=Watch" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github open issues"><img src="https://img.shields.io/github/issues/qorpress/qorpress.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github closed issues"><img src="https://img.shields.io/github/issues-closed/qorpress/qorpress.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github open pr"><img src="https://img.shields.io/github/issues-pr/qorpress/qorpress.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github closed pr"><img src="https://img.shields.io/github/issues-pr-closed/qorpress/qorpress.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github contributors"><img src="https://img.shields.io/github/contributors/qorpress/qorpress.svg" /></a>
    <a href="https://github.com/qorpress/qorpress" alt="github license"><img src="https://img.shields.io/github/license/qorpress/qorpress.svg" /></a>
    <a href="https://travis-ci.com/qorpress/qorpress" alt="travis badge"><img src="https://img.shields.io/travis/qorpress/qorpress.svg" /></a>
</p>

QorPress is a blog engine based on the excellent Qor framework. It aims to be fast and dynamic.

## History
The idea came from the fact that we could not find a blog engine alternative with a back-end/front-end coupled at the same time. 
Hugo is designed for static website as we wanted something allowing to generate dynamic routes with a fast search engine.

## Quick Start

You have basically 2 ways to test QorPress. The first one is to run/build it locally and you will have to install a mysql and a manticore server o your workstation. Either, you can use the docker-compose providing all the required services to run QorPress.

### Local

The requirements are the following:
* Go v1.8+
* MySQL v5.7
* Manticore v3.3+

```shell
# Get QorPress
$ mkdir -p $GOPATH/src/github.com/qorpress
$ git clone --depth=1 --recursive https://github.com/qorpress/qorpress.git
$ cd qorpress

# Setup database
$ mysql -uroot -p
mysql> CREATE DATABASE qorpress;

# Start manticore
$ searchd --config ./.docker/manticore/manticore.conf

# Configure env variables (set the database parameters)
$ cd $GOPATH/src/github.com/qorpress/qorpress
$ mv .env-example .env

# Configure QorPress settings (set the db, ssl, smtp parameters)
$ cd $GOPATH/src/github.com/qorpress/qorpress
$ mv .config/qorpress-example.yml .config/qorpress.yml

# Run Application 
$ go run main.go --compile-templates
$ go run main.go

# Open Browser
$ open http://localhost:7000
$ open https//domain.com # if ssl enabled in qorpress.yml
```

### Docker

The requirements are the following:
* Docker v17+
* Docker-Compose v1.25+

If you don't have Docker/Docker-Compose check **Setup Docker** section
<details>
<summary><b>Setup Docker</b></summary>
<p>

## Docker
macOS: <a href="https://docs.docker.com/docker-for-mac/install/"> https://docs.docker.com/docker-for-mac/install/ </a>

linux: <a href="https://docs.docker.com/install/linux/docker-ce/ubuntu/"> https://docs.docker.com/install/linux/docker-ce/ubuntu/ </a>

## Docker Compose

linux: <a href="https://docs.docker.com/compose/install/"> https://docs.docker.com/compose/install/ </a>
</p>
</details>

```shell
# Get QorPress
$ mkdir -p $GOPATH/src/github.com/qorpress
$ git clone --depth=1 https://github.com/qorpress/qorpress.git
$ cd qorpress

# Run docker containers
$ docker-compose up --build
```

### Generate sample data

based on lorem ipsum texts and fake images
```go
$ cd $GOPATH/src/github.com/qorpress/qorpress
$ go run cmd/lorem/*.go
```

or from kitploit website dump

```go
$ cd $GOPATH/src/github.com/qorpress/qorpress
$ export GITHUB_TOKEN=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
$ go run cmd/kitploit/*.go
```

### Run tests (Pending)

```
$ cd $GOPATH/src/github.com/qorpress/qorpress
$ go test $(go list ./... | grep -v /vendor/ | grep  -v /db/)
```

## Admin Management Interface

[QorPress Example admin configuration](https://github.com/qorpress/qorpress/blob/master/pkg/config/admin/admin.go)

## RESTful API

[QorPress Example API configuration](https://github.com/qorpress/qorpress/blob/master/pkg/config/api/api.go)

Online Example APIs:

* Users: [https://x0rzkov.com/api/users.json](https://x0rzkov.com/api/users.json)
* User 1: [https://x0rzkov.com/api/users/1.json](https://x0rzkov.com/api/users/1.json)
* Posts: [https://x0rzkov.com/api/posts.json](https://x0rzkov.com/api/posts.json)

## Screenshots

### Frontend
#### full post page
![alt text](docs/screenshots/frontend-post_page.png "post page")

### Backend
#### post manager
![alt text](docs/screenshots/backend-list_posts.png "backend list posts")
#### posts edition
![alt text](docs/screenshots/backend-edit_posts.png "backend edit posts")
#### categories manager
![alt text](docs/screenshots/backend-categories.png "backend categories")


## License

Released under the MIT License.

[@GORPRESS](https://twitter.com/gorpress)
