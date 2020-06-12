FROM scratch

COPY server /bin/hello

ENTRYPOINT ["/bin/hello"]