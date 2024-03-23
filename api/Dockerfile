FROM rust:1.70.0-slim-bullseye AS build

# View app name in Cargo.toml
ARG APP_NAME=api
RUN apt update && apt install pkg-config libssl-dev -y
WORKDIR /build

COPY Cargo.lock Cargo.toml ./
RUN sed -i 's/path = "main.rs"/path = "src\/lib.rs"/g' Cargo.toml
RUN mkdir src \
    && echo "fn main() {}" > src/lib.rs \
    && cargo build --release

COPY . .
RUN cargo build --locked --release
RUN cp ./target/release/$APP_NAME /bin/server

FROM debian:bullseye-slim AS final
COPY --from=build /bin/server /bin/
ENV ROCKET_ADDRESS=0.0.0.0
CMD ["/bin/server"]