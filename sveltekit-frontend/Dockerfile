# Stage 1: Build the SvelteKit application
FROM node:20-alpine AS builder

WORKDIR /app

COPY package.json package-lock.json ./

RUN npm ci

COPY . .

RUN npm run build

# Stage 2: Create the final runtime image
FROM node:20-alpine

WORKDIR /app

COPY --from=builder /app/build /app/build
COPY --from=builder /app/package.json /app/package-lock.json ./

RUN npm ci --only=production

EXPOSE 5173

ENV PORT=5173
ENV HOST=0.0.0.0

CMD ["node", "build"]
