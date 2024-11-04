FROM scratch

COPY filestorage /filestorage
COPY config.json /config.json

EXPOSE 8080

CMD ["/filestorage"]