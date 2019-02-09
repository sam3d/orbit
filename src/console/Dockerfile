FROM node:lts AS build
WORKDIR /build
COPY . .
RUN npm install
RUN npm run build

FROM nginx:stable
COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /build/dist /var/www
EXPOSE 6500
