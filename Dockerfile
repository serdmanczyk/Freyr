FROM centurylink/ca-certs

COPY gardenspark /
EXPOSE 8080
ENTRYPOINT ["/gardenspark"]
