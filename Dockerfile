FROM ubuntu:17.04

RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y nodejs nodejs-legacy npm golang-go

RUN apt-get install -y git
RUN git clone https://github.com/anon2anon/esterpad /esterpad

WORKDIR /esterpad
ENV GOPATH /esterpad
ENV PATH /esterpad/bin:$PATH

RUN apt-get install -y protobuf-compiler
RUN go get -u github.com/golang/protobuf/protoc-gen-go

RUN cp config.json.sample config.json
RUN sed -i "s/esterpad:esterpad@localhost/mongo/g" config.json

RUN echo '\n\n\n go get & npm install may take considerable amount of time, please be patient\n\n\n'
RUN make

CMD ["./esterpad"]
