FROM scratch

COPY app /bin/app

ENTRYPOINT ["/bin/app"]
