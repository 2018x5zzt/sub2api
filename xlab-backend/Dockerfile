FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget && rm -rf /var/cache/apk/*

WORKDIR /app
COPY xlab-backend /app/xlab-backend
RUN chmod +x /app/xlab-backend

EXPOSE 8090

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget -q -T 5 -O /dev/null http://127.0.0.1:8090/health || exit 1

CMD ["/app/xlab-backend"]
