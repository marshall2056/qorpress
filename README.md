# QorPress example application

This is an example application to show and explain features of [QOR](http://getqor.com).

Chat Room: [![Join the chat at https://gitter.im/qor/qor](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/qor/qor?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

## Quick Started

### Locally

#### Go version: 1.8+

```shell
# Get example app
$ go get -ugithub.com/qorpress/qorpress-example

# Setup database
$ mysql -uroot -p
mysql> CREATE DATABASE qor_example;

# Run Application
$ cd $GOPATH/src/github.com/qorpress/qorpress-example
$ go run main.go
```

### With Docker

#### Docker version: 

```shell
docker-compose up --build
```

### Generate sample data

```go
$ go run cmd/seeds/*.go
```

### Run tests (Pending)

```
$ go test $(go list ./... | grep -v /vendor/ | grep  -v /db/)
```

## Admin Management Interface

[Qor Example admin configuration](https://github.com/qorpress/qorpress-example/blob/master/config/admin/admin.go)

## RESTful API

[QorPress Example API configuration](https://github.com/qorpress/qorpress-example/blob/master/config/api/api.go)

Online Example APIs:

* Users: [http://demo.getqor.com/api/users.json](http://demo.getqor.com/api/users.json)
* User 1: [http://demo.getqor.com/api/users/1.json](http://demo.getqor.com/api/users/1.json)
* Posts: [http://demo.getqor.com/api/posts.json](http://demo.getqor.com/api/posts.json)

## License

Released under the MIT License.

[@QORSDK](https://twitter.com/qorsdk)
