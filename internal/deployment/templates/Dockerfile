FROM node:22-alpine

WORKDIR /app

COPY package*.json ./

RUN yarn install --production

COPY . .

EXPOSE 3000

CMD ["node", "app.js"]
