# Stage 1: Build the React application
FROM node:22-slim AS builder

WORKDIR /app

# Enable Yarn 2+
RUN corepack enable

COPY package.json yarn.lock .yarnrc.yml* ./
COPY .yarn ./.yarn

RUN yarn install

COPY . .

# Build Arguments
ARG VITE_AUTH0_DOMAIN
ARG VITE_AUTH0_CLIENT_ID
ARG VITE_AUTH0_AUDIENCE
ARG VITE_API_URL

# Set as ENV for build process
ENV VITE_AUTH0_DOMAIN=$VITE_AUTH0_DOMAIN
ENV VITE_AUTH0_CLIENT_ID=$VITE_AUTH0_CLIENT_ID
ENV VITE_AUTH0_AUDIENCE=$VITE_AUTH0_AUDIENCE
ENV VITE_API_URL=$VITE_API_URL

# Build for production
RUN yarn build

# Stage 2: Serve with Nginx
FROM nginx:stable

# Remove default nginx static assets
RUN rm -rf /usr/share/nginx/html/*

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx config if needed (using default for now, but enabling SPA fallback)
# Creating a simple config for SPA support
RUN echo 'server { \
    listen 80; \
    location / { \
    root /usr/share/nginx/html; \
    index index.html index.htm; \
    try_files $uri $uri/ /index.html; \
    } \
    }' > /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
