FROM alpine:latest AS builder
RUN apk --no-cache add build-base

FROM ubuntu AS build1
LABEL sh.recode.vscode.extensions=dbaeumer.vscode-eslint
COPY source1.cpp source.cpp
RUN g++ -o /binary source.cpp

FROM build1 AS build2
COPY source2.cpp source.cpp
RUN g++ -o /binary source.cpp

FROM build2 AS build3
COPY source2.cpp source.cpp
RUN g++ -o /binary source.cpp