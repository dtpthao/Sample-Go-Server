FROM ubuntu:22.04

RUN apt-get update

RUN mkdir /app
WORKDIR /app

ENV WAIT_VERSION 2.9.0
ADD https://github.com/ufoscout/docker-compose-wait/releases/download/$WAIT_VERSION/wait /wait
ADD https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz ./migrate.tar.gz
RUN chmod +x ./wait
RUN tar -xf ./migrate.tar.gz

#copy core app and install required packages
WORKDIR /app
COPY ./db ./db

CMD ./wait && ./migrate -path="$PWD/db" -database="mysql://$DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME" up && ./
