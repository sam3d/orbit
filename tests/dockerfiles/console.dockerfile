FROM node:lts
WORKDIR /app

ENV PORT 5000
EXPOSE 5000

CMD npm install; npm run dev
