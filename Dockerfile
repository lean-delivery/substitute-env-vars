FROM golang as build

WORKDIR /go/src

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o substitute-env-vars main.go


FROM alpine
COPY --from=build /go/src/substitute-env-vars /bin/sev

ENV VAR_NAMES_STORAGE_PATH=""
ENV VAR_NAMES_STORAGE=""

CMD ["sev"]