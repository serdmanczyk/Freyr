FROM centurylink/ca-certs

COPY freyr /
EXPOSE 80
ENTRYPOINT ["/freyr"]
