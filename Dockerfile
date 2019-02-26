# TODO: switch to alpine
FROM ubuntu:18.04

RUN apt-get update
RUN apt-get upgrade -y
RUN apt-get install -y nodejs npm golang-go git

RUN git clone https://github.com/anon2anon/esterpad /esterpad

WORKDIR /esterpad
ENV GOPATH /gopath
ENV PATH /gopath/bin:$PATH

RUN sed -i "s/esterpad:esterpad@localhost/mongo/g" config.yaml

RUN echo '\n\n\n go get & npm install may take considerable amount of time, please be patient\n\n\n'
RUN go get github.com/anon2anon/esterpad/cmd/...
RUN cd web && npm i && npm run build

CMD ["esterpad", "config.yaml"]
