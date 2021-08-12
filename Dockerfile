# Runner Image
FROM alpine
WORKDIR /grpc-mongo-crud/
COPY ./ /grpc-mongo-crud/
EXPOSE 50051

CMD ["/grpc-mongo-crud/bin/server"]