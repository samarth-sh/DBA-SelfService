# Stage 1: Build the SvelteKit application
##
# Build
##
ARG NODE_VERSION=20-alpine
FROM node:${NODE_VERSION} AS builder

ARG BUILD_ENV=production

WORKDIR /app

RUN chown -R node:node /app

# Copy package files first to leverage Docker cache
COPY package.json package-lock.json ./

# Install dependencies
RUN npm install -g npm@latest
USER node
RUN npm install

# Copy the rest of the application
COPY . .

RUN echo "Building for ${BUILD_ENV}"
RUN npm run build:${BUILD_ENV}

# Stage 2: Create the final runtime image
# Archive
FROM node:${NODE_VERSION} as nodejsapp

ENV PORT=8080
USER node
COPY --chown=node:node --from=builder /app/build /app/build
COPY --chown=node:node --from=builder /app/package.json /app/package-lock.json ./

# Install only production dependencies
RUN npm ci --only=production

WORKDIR /app

EXPOSE ${PORT}

# Create entrypoint script
RUN echo -e "#!/bin/sh\n node build/index.js" > entrypoint.sh && \
    chmod +x entrypoint.sh

ENTRYPOINT ["/app/entrypoint.sh"]
