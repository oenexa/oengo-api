# ─────────────────────────────────────────────────────────────────────────────
# Oengo API & Smart Contract Services Dockerfile
# ─────────────────────────────────────────────────────────────────────────────
# Stage 1: Install production dependencies
FROM node:24.21.0-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci --omit=dev

# Stage 2: Minimal runtime
FROM node:24.21.0-alpine

WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3001

# Copy dependencies and application source
COPY --from=builder /app/node_modules ./node_modules
COPY package*.json ./
COPY server.js ./
COPY contracts ./contracts

# Run as non-root user for container security
USER node

EXPOSE 3001

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3001/api/orders || exit 1

CMD ["node", "server.js"]
