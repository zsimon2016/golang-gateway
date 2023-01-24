FROM busybox

RUN mkdir -p /golang_gateway
COPY ./client /golang_gateway/client

RUN mkdir -p /var/log/golang/
RUN touch /var/log/golang/goservices_service_err.log
RUN chmod -R 777 /var/log/golang/goservices_service_err.log

ARG RUN_OPTS=' -nacosAddr 192.168.2.48 -nacosPort 8849'
ENV RUN_OPTS=$RUN_OPTS

EXPOSE 9092

WORKDIR /golang_gateway

ENTRYPOINT [ "sh","-c","/golang_gateway/client $RUN_OPTS > /var/log/golang/goservices_service.log"]
