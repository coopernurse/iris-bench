FROM coopernurse/docker-go:v1

ADD . /usr/local/src/github.com/coopernurse/iris-bench
RUN /usr/local/go/bin/go get gopkg.in/project-iris/iris-go.v1

WORKDIR /usr/local/src/github.com/coopernurse/iris-bench
RUN /usr/local/go/bin/go build -o /usr/local/bin/iris-bench main.go

CMD /usr/local/bin/iris-bench -m server
