# Start from the latest golang base image
FROM golang:alpine AS builder

# Add Maintainer Info
LABEL maintainer="Joseph Keller <jbkeller@gmail.com>"
#Used to make a bash profile we can enter this container with
RUN apk add --no-cache bash
#MakeDirectory
RUN mkdir /app
ADD . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./...

#Make the Scratch Production Image
FROM alpine:latest AS production
#Copy the contents of the builder stage into this app directory
COPY --from=builder /app .
#Expose port 80:80
EXPOSE 8080
#Run the executable,(which is whatever name you gave it in the previous step)
CMD ["./main"]