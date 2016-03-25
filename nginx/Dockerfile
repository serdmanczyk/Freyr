FROM nginx:alpine

COPY ./nginx.conf /etc/nginx/
COPY ./server-cert.pem /etc/nginx/
COPY ./server-key.pem /etc/nginx/
COPY ./static/ /usr/share/nginx/html

