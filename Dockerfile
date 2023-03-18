# This file is used by goreleaser
FROM scratch
ENTRYPOINT ["iracelog-wamp-router"]
COPY iracelog-wamp-router /
COPY routerConfig.yml /
