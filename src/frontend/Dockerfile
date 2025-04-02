# Stage 1: Build the React application
FROM node:20-alpine AS builder

WORKDIR /app

# Copy package.json and package-lock.json (or yarn.lock)
COPY package*.json ./

# Install dependencies
RUN npm install

# Copy the rest of the application source code
COPY . .

# Set API URL build-time argument (can be overridden in docker-compose build args)
ARG REACT_APP_API_URL=/api/v1
ENV REACT_APP_API_URL=$REACT_APP_API_URL

# Build the static files
RUN npm run build

# Stage 2: Serve static files with Nginx
FROM nginx:1.25-alpine

# Copy the built static files from the builder stage
COPY --from=builder /app/build /usr/share/nginx/html

# Optional: Copy custom Nginx configuration for the frontend serving if needed
# COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port 80 (Nginx default)
EXPOSE 80

# Start Nginx
CMD ["nginx", "-g", "daemon off;"]
