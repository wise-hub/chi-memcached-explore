FROM scratch

COPY bin/app bin/app
COPY .env .env

EXPOSE 8080

CMD ["bin/app"]