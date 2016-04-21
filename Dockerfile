FROM centurylink/ca-certs

COPY freyr /
EXPOSE 8080
ENTRYPOINT ["/freyr"]
