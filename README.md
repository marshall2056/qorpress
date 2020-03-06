# QorPress 

QorPress is a blog engine based on the excellent Qor framework. It aims to be fast and dynamic.

## History
The idea came from the fact that we could not find a blog engine alternative with a back-end/front-end coupled at the same time. 
Hugo is designed for static website as we wanted something allowing to generate dynamic routes with a fast search engine.

## Quick Start

### Locally

#### Go version: 1.8+

```shell
# Get QorPress
$ mkdir -p $GOPATH/src/github.com/qorpress
$ git clone --depth=1 https://github.com/qorpress/qorpress.git
$ cd qorpress

# Setup database
$ mysql -uroot -p
mysql> CREATE DATABASE qorpress_example;

# Run Application
$ cd $GOPATH/src/github.com/qorpress/qorpress
$ go run main.go --compile-templates
$ go run main.go
```

### With Docker

#### Docker version: 

```shell
# Get QorPress
$ mkdir -p $GOPATH/src/github.com/qorpress
$ git clone --depth=1 https://github.com/qorpress/qorpress.git
$ cd qorpress

# Run docker containers
docker-compose up --build
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
$ go run cmd/kitploit/*.go
```

### Run tests (Pending)

```
$ cd $GOPATH/src/github.com/qorpress/qorpress
$ go test $(go list ./... | grep -v /vendor/ | grep  -v /db/)
```

## Admin Management Interface

[Qor Example admin configuration](https://github.com/qorpress/qorpress/blob/master/config/admin/admin.go)

## RESTful API

[QorPress Example API configuration](https://github.com/qorpress/qorpress/blob/master/config/api/api.go)

Online Example APIs:

* Users: [http://demo.getqor.com/api/users.json](http://demo.getqor.com/api/users.json)
* User 1: [http://demo.getqor.com/api/users/1.json](http://demo.getqor.com/api/users/1.json)
* Posts: [http://demo.getqor.com/api/posts.json](http://demo.getqor.com/api/posts.json)

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

[@QORSDK](https://twitter.com/qorsdk)
