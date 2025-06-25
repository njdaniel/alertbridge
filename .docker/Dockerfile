FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Copy the pre-built binary from goreleaser
COPY alertbridge .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./alertbridge"] 
