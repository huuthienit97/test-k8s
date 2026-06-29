FROM node:20-alpine
WORKDIR /app

# L4C: image build từ repo root — cần git submodule checkout (libs/is-docker).
COPY libs/is-docker ./libs/is-docker
COPY backend-node/package.json backend-node/index.js ./
RUN npm install ./libs/is-docker --omit=dev

ENV PORT=8080
ENV NODE_ENV=production
USER node
EXPOSE 8080
CMD ["node", "index.js"]
