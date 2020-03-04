FROM golang:apline

WORKDIR /go/src/homef

COPY ./public ./public
COPY ./views ./views

COPY homefapp .

EXPOSE 4000

CMD ["./homefapp"]
