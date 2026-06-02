# Stage 1: Build the React application
FROM node:26-slim AS builder

WORKDIR /app

# Enable Yarn 2+
RUN corepack enable

COPY package.json yarn.lock .yarnrc.yml* ./
COPY .yarn ./.yarn
COPY . .

RUN yarn install --immutable

# Build for production
RUN yarn build

# Stage 2: Serve with Nginx
FROM nginx:stable

# Remove default nginx static assets
RUN rm -rf /usr/share/nginx/html/*

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Generate runtime config.js from container environment on startup.
COPY docker-entrypoint.d/40-generate-config.sh /docker-entrypoint.d/40-generate-config.sh
RUN chmod +x /docker-entrypoint.d/40-generate-config.sh


# Copy custom nginx config for API proxying and SPA fallback
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
