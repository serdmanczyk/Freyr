FROM centurylink/ca-certs

COPY ./static/index.html /static/index.html
COPY ./static/build/bundle.js /static/build/bundle.js
COPY ./static/js/jquery.js /static/js/js/jquery.js
COPY ./static/js/bootstrap.min.js /static/js/bootstrap.min.js
COPY ./static/js/jquery.easing.min.js /static/js/jquery.easing.min.js
COPY ./static/js/grayscale.js /static/js/grayscale.js
COPY ./static/css/ /static/css/
COPY ./static/font-awesome/ /static/font-awesome/
COPY ./static/img/ /static/img/
COPY freyr /
EXPOSE 80
ENTRYPOINT ["/freyr"]
