FROM nginx:alpine

COPY ./conf/nginx.conf /etc/nginx/
COPY ./conf/server-cert.pem /etc/nginx/
COPY ./conf/server-key.pem /etc/nginx/

RUN mkdir -p /usr/share/nginx/html/build
RUN mkdir -p /usr/share/nginx/html/css
RUN mkdir -p /usr/share/nginx/html/font-awesome
RUN mkdir -p /usr/share/nginx/html/img
COPY ./static/index.html /usr/share/nginx/html
COPY ./static/build/bundle.js /usr/share/nginx/html/build
COPY ./static/css/ /usr/share/nginx/html/css/
COPY ./static/font-awesome/ /usr/share/nginx/html/font-awesome/
COPY ./static/img/ /usr/share/nginx/html/img/
