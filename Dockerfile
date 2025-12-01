FROM alpine:latest


RUN apk add --no-cache bind-tools ca-certificates libc6-compat


COPY leaderboard /app/leaderboard
COPY wait-for-dns.sh /app/wait-for-dns.sh


RUN chmod +x /app/leaderboard /app/wait-for-dns.sh

WORKDIR /app


CMD ["/app/wait-for-dns.sh", "postgres", "/app/leaderboard"]