FROM node:24-alpine
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci --ignore-scripts
COPY server.mjs ./
USER node
CMD ["node", "server.mjs"]
