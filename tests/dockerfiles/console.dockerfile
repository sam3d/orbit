FROM node:lts
WORKDIR /app

ENV PORT 6500
EXPOSE 6500

CMD npm install; npm run dev
