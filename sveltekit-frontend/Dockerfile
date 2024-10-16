# Stage 1: Build the SvelteKit application
##
# Build
##
ARG NODE_VERSION=20-alpine
FROM node:${NODE_VERSION} AS builder

ARG BUILD_ENV

WORKDIR /app
COPY . .

# Change ownership of /app to node user
RUN chown -R node:node /app

USER node

RUN echo BUILD_ENV is ${BUILD_ENV}
RUN npm install
RUN npm run build:$BUILD_ENV

# Stage 2: Create the final runtime image
# Archive
FROM node:${NODE_VERSION} AS nodejsapp

ENV PORT=8080
USER node
COPY --chown=node:node --from=builder /app /app

EXPOSE 8080

WORKDIR /app

EXPOSE ${PORT}

# Create entrypoint script
RUN echo -e "#!/bin/sh\n node build/index.js" > entrypoint.sh && \
    chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
