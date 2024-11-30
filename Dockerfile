FROM scratch

COPY filestorage /filestorage

EXPOSE 8080

CMD ["/filestorage"]