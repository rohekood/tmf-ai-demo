FROM node:22-alpine

WORKDIR /app

# Copy configuration files
COPY package.json yarn.lock .yarnrc.yml* ./
COPY .yarn ./.yarn

# Install dependencies
RUN yarn install

# Copy source code
COPY . .

EXPOSE 5173

CMD ["yarn", "dev", "--host"]
