FROM nginx:alpine

COPY ./conf/nginx.conf /etc/nginx/
COPY ./conf/server-cert.pem /etc/nginx/
COPY ./conf/server-key.pem /etc/nginx/

RUN mkdir -p /usr/share/nginx/html/build
RUN mkdir -p /usr/share/nginx/html/css
COPY ./static/index.html /usr/share/nginx/html
COPY ./static/build/bundle.js /usr/share/nginx/html
COPY ./static/css/ /usr/share/nginx/html/css/
