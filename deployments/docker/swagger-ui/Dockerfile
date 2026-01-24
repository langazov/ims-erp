FROM swaggerapi/swagger-ui:latest

LABEL maintainer="erp@example.com"
LABEL description="ERP System API Documentation"

# Copy OpenAPI spec
COPY api/openapi.yaml /openapi/
COPY docs/swagger/index.html /usr/share/nginx/html/index.html

# Expose port
EXPOSE 80

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:80/ || exit 1

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
