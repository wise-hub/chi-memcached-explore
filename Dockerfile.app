FROM scratch

COPY bin/app /app

EXPOSE 8080

CMD ["/app"]